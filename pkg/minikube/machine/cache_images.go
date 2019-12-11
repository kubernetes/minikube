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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// loadRoot is where images should be loaded from within the guest VM
var loadRoot = path.Join(vmpath.GuestPersistentDir, "images")

var getWindowsVolumeName = getWindowsVolumeNameCmd

// loadImageLock is used to serialize image loads to avoid overloading the guest VM
var loadImageLock sync.Mutex

// CacheImagesForBootstrapper will cache images for a bootstrapper
func CacheImagesForBootstrapper(imageRepository string, version string, clusterBootstrapper string) error {
	images, err := bootstrapper.GetCachedImageList(imageRepository, version, clusterBootstrapper)
	if err != nil {
		return errors.Wrap(err, "cached images list")
	}

	if err := CacheImagesToTar(images, constants.ImageCacheDir); err != nil {
		return errors.Wrapf(err, "Caching images for %s", clusterBootstrapper)
	}

	return nil
}

// CacheImagesToTar will cache images on the host
//
// The cache directory currently caches images using the imagename_tag
// For example, k8s.gcr.io/kube-addon-manager:v6.5 would be
// stored at $CACHE_DIR/k8s.gcr.io/kube-addon-manager_v6.5
func CacheImagesToTar(images []string, cacheDir string) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			dst := filepath.Join(cacheDir, image)
			dst = sanitizeCacheDir(dst)
			if err := cacheImageToTarFile(image, dst); err != nil {
				glog.Errorf("CacheImage %s -> %s failed: %v", image, dst, err)
				return errors.Wrapf(err, "caching image %s", dst)
			}
			glog.Infof("CacheImage %s -> %s succeeded", image, dst)
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
func LoadImages(cc *config.MachineConfig, runner command.Runner, images []string, cacheDir string) error {
	glog.Infof("LoadImages start: %s", images)
	defer glog.Infof("LoadImages end")
	var g errgroup.Group
	cr, err := cruntime.New(cruntime.Config{Type: cc.ContainerRuntime, Runner: runner})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	for _, image := range images {
		image := image
		g.Go(func() error {
			err := needsTransfer(image, cr)
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
func needsTransfer(image string, cr cruntime.Manager) error {
	ref, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return errors.Wrap(err, "parse ref")
	}

	img, err := retrieveImage(ref)
	if err != nil {
		return errors.Wrap(err, "retrieve")
	}

	cf, err := img.ConfigName()
	if err != nil {
		return errors.Wrap(err, "image hash")
	}

	if !cr.ImageExists(image, cf.Hex) {
		return fmt.Errorf("%q does not exist at hash %q in container runtime", image, cf.Hex)
	}
	return nil
}

// CacheAndLoadImages caches and loads images to all profiles
func CacheAndLoadImages(images []string) error {
	if err := CacheImagesToTar(images, constants.ImageCacheDir); err != nil {
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
		status, err := cluster.GetHostStatus(api, pName)
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

// # ParseReference cannot have a : in the directory path
func sanitizeCacheDir(image string) string {
	if runtime.GOOS == "windows" && hasWindowsDriveLetter(image) {
		// not sanitize Windows drive letter.
		s := image[:2] + strings.Replace(image[2:], ":", "_", -1)
		glog.Infof("windows sanitize: %s -> %s", image, s)
		return s
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

// transferAndLoadImage transfers and loads a single image from the cache
func transferAndLoadImage(cr command.Runner, k8s config.KubernetesConfig, imgName string, cacheDir string) error {
	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: cr})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}
	src := filepath.Join(cacheDir, imgName)
	src = sanitizeCacheDir(src)
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

// cacheImageToTarFile caches an image
func cacheImageToTarFile(image, dst string) error {
	start := time.Now()
	glog.Infof("CacheImage: %s -> %s", image, dst)
	defer func() {
		glog.Infof("CacheImage: %s -> %s completed in %s", image, dst, time.Since(start))
	}()

	if _, err := os.Stat(dst); err == nil {
		glog.Infof("%s exists", dst)
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

	img, err := retrieveImage(ref)
	if err != nil {
		glog.Warningf("unable to retrieve image: %v", err)
	}

	glog.Infoln("OPENING: ", dstPath)
	f, err := ioutil.TempFile(filepath.Dir(dstPath), filepath.Base(dstPath)+".*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		// If we left behind a temp file, remove it.
		_, err := os.Stat(f.Name())
		if err == nil {
			os.Remove(f.Name())
			if err != nil {
				glog.Warningf("Failed to clean up the temp file %s: %v", f.Name(), err)
			}
		}
	}()
	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		return errors.Wrap(err, "newtag")
	}
	err = tarball.Write(tag, img, &tarball.WriteOptions{}, f)
	if err != nil {
		return errors.Wrap(err, "write")
	}
	err = f.Close()
	if err != nil {
		return errors.Wrap(err, "close")
	}
	err = os.Rename(f.Name(), dstPath)
	if err != nil {
		return errors.Wrap(err, "rename")
	}
	glog.Infof("%s exists", dst)
	return nil
}

func retrieveImage(ref name.Reference) (v1.Image, error) {
	glog.Infof("retrieving image: %+v", ref)
	img, err := daemon.Image(ref)
	if err == nil {
		glog.Infof("found %s locally: %+v", ref.Name(), img)
		return img, nil
	}
	// reference does not exist in the local daemon
	if err != nil {
		glog.Infof("daemon lookup for %+v: %v", ref, err)
	}

	img, err = remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err == nil {
		return img, nil
	}

	glog.Warningf("authn lookup for %+v (trying anon): %+v", ref, err)
	img, err = remote.Image(ref)
	return img, err
}
