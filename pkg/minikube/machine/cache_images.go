/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machine

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/image"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// loadRoot is where images should be loaded from within the guest VM
var loadRoot = path.Join(vmpath.GuestPersistentDir, "images")

// loadImageLock is used to serialize image loads to avoid overloading the guest VM
var loadImageLock sync.Mutex

// CacheImagesForBootstrapper will cache images for a bootstrapper
func CacheImagesForBootstrapper(imageRepository string, version string, clusterBootstrapper string) error {
	images, err := bootstrapper.GetCachedImageList(imageRepository, version, clusterBootstrapper)
	if err != nil {
		return errors.Wrap(err, "cached images list")
	}

	if err := image.SaveToDir(images, constants.ImageCacheDir); err != nil {
		return errors.Wrapf(err, "Caching images for %s", clusterBootstrapper)
	}

	return nil
}

// LoadImages loads previously cached images into the container runtime
func LoadImages(cc *config.ClusterConfig, runner command.Runner, images []string, cacheDir string) error {
	glog.Infof("LoadImages start: %s", images)
	start := time.Now()

	defer func() {
		glog.Infof("LoadImages completed in %s", time.Since(start))
	}()

	var g errgroup.Group
	cr, err := cruntime.New(cruntime.Config{Type: cc.KubernetesConfig.ContainerRuntime, Runner: runner})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	imgClient, err := client.NewClientWithOpts(client.FromEnv) // image client
	if err != nil {
		glog.Infof("couldn't get a local image daemon which might be ok: %v", err)
		imgClient = nil
	}

	for _, image := range images {
		image := image
		g.Go(func() error {
			err := needsTransfer(imgClient, image, cr)
			if err == nil {
				return nil
			}
			glog.Infof("%q needs transfer: %v", image, err)
			return transferAndLoadImage(runner, cc.KubernetesConfig, image, cacheDir)
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "loading cached images")
	}
	glog.Infoln("Successfully loaded all cached images")
	return nil
}

// needsTransfer returns an error if an image needs to be retransfered
func needsTransfer(imgClient *client.Client, imgName string, cr cruntime.Manager) error {
	imgDgst := ""         // for instance sha256:7c92a2c6bbcb6b6beff92d0a940779769c2477b807c202954c537e2e0deb9bed
	if imgClient != nil { // if possible try to get img digest from Client lib which is 4s faster.
		imgDgst = image.DigestByDockerLib(imgClient, imgName)
		if imgDgst != "" {
			if !cr.ImageExists(imgName, imgDgst) {
				return fmt.Errorf("%q does not exist at hash %q in container runtime", imgName, imgDgst)
			}
			return nil
		}
	}
	// if not found with method above try go-container lib (which is 4s slower)
	imgDgst = image.DigestByGoLib(imgName)
	if imgDgst == "" {
		return fmt.Errorf("got empty img digest %q for %s", imgDgst, imgName)
	}
	if !cr.ImageExists(imgName, imgDgst) {
		return fmt.Errorf("%q does not exist at hash %q in container runtime", imgName, imgDgst)
	}
	return nil
}

// CacheAndLoadImages caches and loads images to all profiles
func CacheAndLoadImages(images []string) error {
	if err := image.SaveToDir(images, constants.ImageCacheDir); err != nil {
		return err
	}
	api, err := NewAPIClient()
	if err != nil {
		return err
	}
	defer api.Close()
	profiles, _, err := config.ListProfiles() // need to load image to all profiles
	if err != nil {
		return errors.Wrap(err, "list profiles")
	}
	for _, p := range profiles { // loading images to all running profiles
		pName := p.Name // capture the loop variable
		status, err := GetHostStatus(api, pName)
		if err != nil {
			glog.Warningf("skipping loading cache for profile %s", pName)
			glog.Errorf("error getting status for %s: %v", pName, err)
			continue // try next machine
		}
		if status == state.Running.String() { // the not running hosts will load on next start
			h, err := api.Load(pName)
			if err != nil {
				return err
			}
			cr, err := CommandRunner(h)
			if err != nil {
				return err
			}
			c, err := config.Load(pName)
			if err != nil {
				return err
			}
			err = LoadImages(c, cr, images, constants.ImageCacheDir)
			if err != nil {
				glog.Warningf("Failed to load cached images for profile %s. make sure the profile is running. %v", pName, err)
			}
		}
	}
	return err
}

// transferAndLoadImage transfers and loads a single image from the cache
func transferAndLoadImage(cr command.Runner, k8s config.KubernetesConfig, imgName string, cacheDir string) error {
	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: cr})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}
	src := filepath.Join(cacheDir, imgName)
	src = localpath.SanitizeCacheDir(src)
	glog.Infof("Loading image from cache: %s", src)
	filename := filepath.Base(src)
	if _, err := os.Stat(src); err != nil {
		return err
	}
	dst := path.Join(loadRoot, filename)
	f, err := assets.NewFileAsset(src, loadRoot, filename, "0644")
	if err != nil {
		return errors.Wrapf(err, "creating copyable file asset: %s", filename)
	}
	if err := cr.Copy(f); err != nil {
		return errors.Wrap(err, "transferring cached image")
	}

	loadImageLock.Lock()
	defer loadImageLock.Unlock()

	err = r.LoadImage(dst)
	if err != nil {
		return errors.Wrapf(err, "%s load %s", r.Name(), dst)
	}

	glog.Infof("Transferred and loaded %s from cache", src)
	return nil
}
