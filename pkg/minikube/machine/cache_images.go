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
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/sshutil"
)

const tempLoadDir = "/tmp"

var getWindowsVolumeName = getWindowsVolumeNameCmd

// loadImageLock is used to serialize image loads to avoid overloading the guest VM
var loadImageLock sync.Mutex

// CacheImagesForBootstrapper will cache images for a bootstrapper
func CacheImagesForBootstrapper(imageRepository string, version string, clusterBootstrapper string) error {
	images := bootstrapper.GetCachedImageList(imageRepository, version, clusterBootstrapper)

	if err := CacheImages(images, constants.ImageCacheDir); err != nil {
		return errors.Wrapf(err, "Caching images for %s", clusterBootstrapper)
	}

	return nil
}

// CacheImages will cache images on the host
//
// The cache directory currently caches images using the imagename_tag
// For example, k8s.gcr.io/kube-addon-manager:v6.5 would be
// stored at $CACHE_DIR/k8s.gcr.io/kube-addon-manager_v6.5
func CacheImages(images []string, cacheDir string) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			dst := filepath.Join(cacheDir, image)
			dst = sanitizeCacheDir(dst)
			if err := CacheImage(image, dst); err != nil {
				return errors.Wrapf(err, "caching image %s", dst)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "caching images")
	}
	glog.Infoln("Successfully cached all images.")
	return nil
}

// LoadImages loads previously cached images into the container runtime
func LoadImages(cmd bootstrapper.CommandRunner, images []string, cacheDir string) error {
	var g errgroup.Group
	// Load profile cluster config from file
	cc, err := config.Load()
	if err != nil && !os.IsNotExist(err) {
		glog.Errorln("Error loading profile config: ", err)
	}
	for _, image := range images {
		image := image
		g.Go(func() error {
			src := filepath.Join(cacheDir, image)
			src = sanitizeCacheDir(src)
			if err := loadImageFromCache(cmd, cc.KubernetesConfig, src); err != nil {
				glog.Warningf("Failed to load %s: %v", src, err)
				return errors.Wrapf(err, "loading image %s", src)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "loading cached images")
	}
	glog.Infoln("Successfully loaded all cached images.")
	return nil
}

// CacheAndLoadImages caches and loads images
func CacheAndLoadImages(images []string) error {
	if err := CacheImages(images, constants.ImageCacheDir); err != nil {
		return err
	}
	api, err := NewAPIClient()
	if err != nil {
		return err
	}
	defer api.Close()
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return err
	}

	client, err := sshutil.NewSSHClient(h.Driver)
	if err != nil {
		return err
	}
	runner := bootstrapper.NewSSHRunner(client)
	return LoadImages(runner, images, constants.ImageCacheDir)
}

// # ParseReference cannot have a : in the directory path
func sanitizeCacheDir(image string) string {
	if runtime.GOOS == "windows" && hasWindowsDriveLetter(image) {
		// not sanitize Windows drive letter.
		return image[:2] + strings.Replace(image[2:], ":", "_", -1)
	}
	return strings.Replace(image, ":", "_", -1)
}

func hasWindowsDriveLetter(s string) bool {
	if len(s) < 3 {
		return false
	}

	drive := s[:3]
	for _, b := range "CDEFGHIJKLMNOPQRSTUVWXYZABcdefghijklmnopqrstuvwxyzab" {
		if d := string(b) + ":"; drive == d+`\` || drive == d+`/` {
			return true
		}
	}

	return false
}

// Replace a drive letter to a volume name.
func replaceWinDriveLetterToVolumeName(s string) (string, error) {
	vname, err := getWindowsVolumeName(s[:1])
	if err != nil {
		return "", err
	}
	path := vname + s[3:]

	return path, nil
}

func getWindowsVolumeNameCmd(d string) (string, error) {
	cmd := exec.Command("wmic", "volume", "where", "DriveLetter = '"+d+":'", "get", "DeviceID")

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	outs := strings.Split(strings.Replace(string(stdout), "\r", "", -1), "\n")

	var vname string
	for _, l := range outs {
		s := strings.TrimSpace(l)
		if strings.HasPrefix(s, `\\?\Volume{`) && strings.HasSuffix(s, `}\`) {
			vname = s
			break
		}
	}

	if vname == "" {
		return "", errors.New("failed to get a volume GUID")
	}

	return vname, nil
}

// loadImageFromCache loads a single image from the cache
func loadImageFromCache(cr bootstrapper.CommandRunner, k8s config.KubernetesConfig, src string) error {
	glog.Infof("Loading image from cache: %s", src)
	filename := filepath.Base(src)
	if _, err := os.Stat(src); err != nil {
		return err
	}
	dst := path.Join(tempLoadDir, filename)
	f, err := assets.NewFileAsset(src, tempLoadDir, filename, "0777")
	if err != nil {
		return errors.Wrapf(err, "creating copyable file asset: %s", filename)
	}
	if err := cr.Copy(f); err != nil {
		return errors.Wrap(err, "transferring cached image")
	}

	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: cr})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}
	loadImageLock.Lock()
	defer loadImageLock.Unlock()

	err = r.LoadImage(dst)
	if err != nil {
		return errors.Wrapf(err, "%s load %s", r.Name(), dst)
	}

	if err := cr.Run("sudo rm -rf " + dst); err != nil {
		return errors.Wrap(err, "deleting temp docker image location")
	}
	glog.Infof("Successfully loaded image %s from cache", src)
	return nil
}

// DeleteFromImageCacheDir deletes images from the cache
func DeleteFromImageCacheDir(images []string) error {
	for _, image := range images {
		path := filepath.Join(constants.ImageCacheDir, image)
		path = sanitizeCacheDir(path)
		glog.Infoln("Deleting image in cache at ", path)
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return cleanImageCacheDir()
}

func cleanImageCacheDir() error {
	err := filepath.Walk(constants.ImageCacheDir, func(path string, info os.FileInfo, err error) error {
		// If error is not nil, it's because the path was already deleted and doesn't exist
		// Move on to next path
		if err != nil {
			return nil
		}
		// Check if path is directory
		if !info.IsDir() {
			return nil
		}
		// If directory is empty, delete it
		entries, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			if err = os.Remove(path); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func getDstPath(dst string) (string, error) {
	if runtime.GOOS == "windows" && hasWindowsDriveLetter(dst) {
		// ParseReference does not support a Windows drive letter.
		// Therefore, will replace the drive letter to a volume name.
		var err error
		if dst, err = replaceWinDriveLetterToVolumeName(dst); err != nil {
			return "", errors.Wrap(err, "parsing docker archive dst ref: replace a Win drive letter to a volume name")
		}
	}

	return dst, nil
}

// CacheImage caches an image
func CacheImage(image, dst string) error {
	// There are go-containerregistry calls here that result in
	// ugly log messages getting printed to stdout. Capture
	// stdout instead and writing it to info.
	r, w, err := os.Pipe()
	if err != nil {
		return errors.Wrap(err, "opening writing buffer")
	}
	log.SetOutput(w)
	defer func() {
		log.SetOutput(os.Stdout)
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			glog.Errorf("output copy failed: %v", err)
		}
		glog.Infof(buf.String())
	}()

	glog.Infof("Attempting to cache image: %s at %s\n", image, dst)
	if _, err := os.Stat(dst); err == nil {
		return nil
	}

	dstPath, err := getDstPath(dst)
	if err != nil {
		return errors.Wrap(err, "getting destination path")
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
		return errors.Wrapf(err, "making cache image directory: %s", dst)
	}

	ref, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return errors.Wrap(err, "creating docker image name")
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return errors.Wrap(err, "fetching remote image")
	}

	glog.Infoln("OPENING: ", dstPath)
	f, err := ioutil.TempFile(filepath.Dir(dstPath), filepath.Base(dstPath)+".*.tmp")
	if err != nil {
		return err
	}
	err = tarball.Write(ref, img, f)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	err = os.Rename(f.Name(), dstPath)
	if err != nil {
		return err
	}
	return nil
}
