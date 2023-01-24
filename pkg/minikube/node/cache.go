/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package node

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"k8s.io/minikube/pkg/minikube/detect"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/image"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	cacheImages         = "cache-images"
	cacheImageConfigKey = "cache"
)

// BeginCacheKubernetesImages caches images required for Kubernetes version in the background
func beginCacheKubernetesImages(g *errgroup.Group, imageRepository string, k8sVersion string, cRuntime string, driverName string) {
	// TODO: remove imageRepository check once #7695 is fixed
	if imageRepository == "" && download.PreloadExists(k8sVersion, cRuntime, driverName) {
		klog.Info("Caching tarball of preloaded images")
		err := download.Preload(k8sVersion, cRuntime, driverName)
		if err == nil {
			klog.Infof("Finished verifying existence of preloaded tar for  %s on %s", k8sVersion, cRuntime)
			return // don't cache individual images if preload is successful.
		}
		klog.Warningf("Error downloading preloaded artifacts will continue without preload: %v", err)
	}

	if !viper.GetBool(cacheImages) {
		return
	}

	g.Go(func() error {
		return machine.CacheImagesForBootstrapper(imageRepository, k8sVersion, viper.GetString(cmdcfg.Bootstrapper))
	})
}

// handleDownloadOnly caches appropariate binaries and images
func handleDownloadOnly(cacheGroup, kicGroup *errgroup.Group, k8sVersion, containerRuntime, driverName string) {
	// If --download-only, complete the remaining downloads and exit.
	if !viper.GetBool("download-only") {
		return
	}

	binariesURL := viper.GetString("binary-mirror")
	if err := doCacheBinaries(k8sVersion, containerRuntime, driverName, binariesURL); err != nil {
		exit.Error(reason.InetCacheBinaries, "Failed to cache binaries", err)
	}
	if _, err := CacheKubectlBinary(k8sVersion, binariesURL); err != nil {
		exit.Error(reason.InetCacheKubectl, "Failed to cache kubectl", err)
	}
	waitCacheRequiredImages(cacheGroup)
	waitDownloadKicBaseImage(kicGroup)
	if err := saveImagesToTarFromConfig(); err != nil {
		exit.Error(reason.InetCacheTar, "Failed to cache images to tar", err)
	}
	out.Step(style.Check, "Download complete!")
	os.Exit(0)
}

// CacheKubectlBinary caches the kubectl binary
func CacheKubectlBinary(k8sVersion, binaryURL string) (string, error) {
	binary := "kubectl"
	if runtime.GOOS == "windows" {
		binary = "kubectl.exe"
	}

	return download.Binary(binary, k8sVersion, runtime.GOOS, detect.EffectiveArch(), binaryURL)
}

// doCacheBinaries caches Kubernetes binaries in the foreground
func doCacheBinaries(k8sVersion, containerRuntime, driverName, binariesURL string) error {
	existingBinaries := constants.KubernetesReleaseBinaries
	if !download.PreloadExists(k8sVersion, containerRuntime, driverName) {
		existingBinaries = nil
	}
	return machine.CacheBinariesForBootstrapper(k8sVersion, viper.GetString(cmdcfg.Bootstrapper), existingBinaries, binariesURL)
}

// beginDownloadKicBaseImage downloads the kic image
func beginDownloadKicBaseImage(g *errgroup.Group, cc *config.ClusterConfig, downloadOnly bool) {

	klog.Infof("Beginning downloading kic base image for %s with %s", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
	register.Reg.SetStep(register.PullingBaseImage)
	g.Go(func() error {
		baseImg := cc.KicBaseImage
		if baseImg == kic.BaseImage && len(cc.KubernetesConfig.ImageRepository) != 0 {
			baseImg = updateKicImageRepo(baseImg, cc.KubernetesConfig.ImageRepository)
			cc.KicBaseImage = baseImg
		}
		var finalImg string
		// If we end up using a fallback image, notify the user
		defer func() {
			if finalImg != "" {
				cc.KicBaseImage = finalImg
				if image.Tag(finalImg) != image.Tag(baseImg) {
					out.WarningT(fmt.Sprintf("minikube was unable to download %s, but successfully downloaded %s as a fallback image", image.Tag(baseImg), image.Tag(finalImg)))
				}
			}
		}()
		for _, img := range append([]string{baseImg}, kic.FallbackImages...) {

			// 1. check if image is already in minikube cache
			cached, err := isImageAlreadyCached(img)
			if err != nil {
				return err
			}

			// if we don't have the specified image
			// we try to pull it from remote, to the minikube cache
			// if we have it already.. we're good
			// takes into account the --download-only flag
			if !cached || downloadOnly {
				out.Step(style.Pulling, "Pulling base image to minikube cache ...")
				klog.Infof("Downloading %s to minikube cache", img)
				err = pullImageToMinikubeCache(img)
				if err == nil && downloadOnly {
					return nil
				} else if err != nil {
					return err
				}
			}

			// 2. We check that the image is loaded inside the kicDriver
			cd, err := getContentDigestFromTarball(img)
			if err != nil {
				return err
			}
			stored, err := isImageInKicDriver(cc.Driver, cd)
			if err != nil {
				return err
			}
			if stored {
				klog.Infof("%s already present in KicDriver", img)
				finalImg = img
				return nil
			}

			out.Step(style.Waiting, "Loading KicDriver with base image ...")
			// if we don't have the cached image in KicDriver.. we're loading it
			if err := download.CacheToKicDriver(cc.Driver, img); err == nil {
				klog.Infof("successfully loaded and using %s from cached tarball", img)
				return nil
			}
			klog.Infof("failed to download %s, will try fallback image if available: %v", img, err)
		}
		return fmt.Errorf("failed to download kic base image or any fallback image")
	})
}

// waitDownloadKicBaseImage blocks until the base image for KIC is downloaded.
func waitDownloadKicBaseImage(g *errgroup.Group) {
	if err := g.Wait(); err != nil {
		if err != nil {
			if errors.Is(err, image.ErrGithubNeedsLogin) {
				klog.Warningf("Error downloading kic artifacts: %v", err)
				out.ErrT(style.Connectivity, "Unfortunately, could not download the base image {{.image_name}} ", out.V{"image_name": image.Tag(kic.BaseImage)})
				out.WarningT("In order to use the fall back image, you need to log in to the github packages registry")
				out.Styled(style.Documentation, `Please visit the following link for documentation around this: 
	https://help.github.com/en/packages/using-github-packages-with-your-projects-ecosystem/configuring-docker-for-use-with-github-packages#authenticating-to-github-packages
`)
			}
			if errors.Is(err, image.ErrGithubNeedsLogin) || errors.Is(err, image.ErrNeedsLogin) {
				exit.Message(reason.Usage, `Please either authenticate to the registry or use --base-image flag to use a different registry.`)
			} else {
				klog.Errorln("Error downloading kic artifacts: ", err)
			}

		}
	}
	klog.Info("Successfully downloaded all kic artifacts")
}

// waitCacheRequiredImages blocks until the required images are all cached.
func waitCacheRequiredImages(g *errgroup.Group) {
	if !viper.GetBool(cacheImages) {
		return
	}
	if err := g.Wait(); err != nil {
		klog.Errorln("Error caching images: ", err)
	}
}

// saveImagesToTarFromConfig saves images to tar in cache which specified in config file.
// currently only used by download-only option
func saveImagesToTarFromConfig() error {
	images, err := imagesInConfigFile()
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return nil
	}
	return image.SaveToDir(images, detect.ImageCacheDir(), false)
}

// CacheAndLoadImagesInConfig loads the images currently in the config file
// called by 'start' and 'cache reload' commands.
func CacheAndLoadImagesInConfig(profiles []*config.Profile) error {
	images, err := imagesInConfigFile()
	if err != nil {
		return errors.Wrap(err, "images")
	}
	if len(images) == 0 {
		return nil
	}
	return machine.CacheAndLoadImages(images, profiles, false)
}

func imagesInConfigFile() ([]string, error) {
	configFile, err := config.ReadConfig(localpath.ConfigFile())
	if err != nil {
		return nil, errors.Wrap(err, "read")
	}
	if values, ok := configFile[cacheImageConfigKey]; ok {
		var images []string
		for key := range values.(map[string]interface{}) {
			images = append(images, key)
		}
		return images, nil
	}
	return []string{}, nil
}

func updateKicImageRepo(imgName string, repo string) string {
	image := strings.TrimPrefix(imgName, "gcr.io/")
	if repo == constants.AliyunMirror {
		// for aliyun registry must strip namespace from image name, e.g.
		//   registry.cn-hangzhou.aliyuncs.com/google_containers/k8s-minikube/kicbase:v0.0.25 will not work
		//   registry.cn-hangzhou.aliyuncs.com/google_containers/kicbase:v0.0.25 does work
		image = strings.TrimPrefix(image, "k8s-minikube/")
	}
	return path.Join(repo, image)
}

// isImageAlreadyCached
// given an IMG(img[:tag[@digest]]),
// it sanitizes the string as a path to minikube cache
// and looks for the referenced tarball
func isImageAlreadyCached(img string) (bool, error) {
	fname := download.ImagePathInCache(img)
	exists, err := func(p string) (bool, error) {
		_, err := os.Stat(p)
		if err == nil {
			return true, nil
		}

		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}(fname)
	if err != nil {
		return false, err
	}

	return exists, err
}

// pullImageToMinikubeCache
// given an IMG, it saves it in .tar format inside the minikube cache
// if successful, it reads the contentDigest from the .tar file
// and saves all the information as a repositories.json entry
func pullImageToMinikubeCache(img string) error {
	return download.ImageToCache(img)
}

// isImageInKicDriver
// takes a contentDigest CD as a parameter,
// checks if the corresponding image is loaded inside the kicDriver
// TODO: perhaps we should call the kic pkg, that in turn should call the oci pkg?
func isImageInKicDriver(ociBin, cd string) (bool, error) {
	return oci.IsImageLoaded(ociBin, cd)
}

// getContentDigestFromTarball
// takes an IMG as a parameter, finds the related tarball.
// walks tarball header for manifest.json and return the Config hash
func getContentDigestFromTarball(img string) (string, error) {
	var manifest = "manifest.json"

	tarPath := download.ImagePathInCache(img)
	fd, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	tr := tar.NewReader(fd)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return "", fmt.Errorf("reached EOF in %s while looking for manifest", tarPath)
		}
		if err != nil {
			return "", err
		}

		if hdr.Name == manifest {
			data, err := io.ReadAll(tr)
			if err != nil {
				return "", err
			}

			imgCfg := []struct {
				Cfg string `json:"Config"`
			}{}
			err = json.Unmarshal(data, &imgCfg)
			if err != nil {
				return "", err
			}

			cd := strings.TrimPrefix(imgCfg[0].Cfg, "sha256:")
			cd = strings.TrimSuffix(cd, ".json")
			return cd, nil
		}
	}
}
