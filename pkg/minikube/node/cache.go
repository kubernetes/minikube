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
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
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
func beginCacheKubernetesImages(g *errgroup.Group, imageRepository string, k8sVersion string, cRuntime string) {
	// TODO: remove imageRepository check once #7695 is fixed
	if imageRepository == "" && download.PreloadExists(k8sVersion, cRuntime) {
		klog.Info("Caching tarball of preloaded images")
		err := download.Preload(k8sVersion, cRuntime)
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

// HandleDownloadOnly caches appropariate binaries and images
func handleDownloadOnly(cacheGroup, kicGroup *errgroup.Group, k8sVersion string) {
	// If --download-only, complete the remaining downloads and exit.
	if !viper.GetBool("download-only") {
		return
	}
	if err := doCacheBinaries(k8sVersion); err != nil {
		exit.Error(reason.InetCacheBinaries, "Failed to cache binaries", err)
	}
	if _, err := CacheKubectlBinary(k8sVersion); err != nil {
		exit.Error(reason.InetCacheKubectl, "Failed to cache kubectl", err)
	}
	waitCacheRequiredImages(cacheGroup)
	waitDownloadKicBaseImage(kicGroup)
	if err := saveImagesToTarFromConfig(); err != nil {
		exit.Error(reason.InetCacheTar, "Failed to cache images to tar", err)
	}
	out.T(style.Check, "Download complete!")
	os.Exit(0)
}

// CacheKubectlBinary caches the kubectl binary
func CacheKubectlBinary(k8sVersion string) (string, error) {
	binary := "kubectl"
	if runtime.GOOS == "windows" {
		binary = "kubectl.exe"
	}

	return download.Binary(binary, k8sVersion, runtime.GOOS, runtime.GOARCH)
}

// doCacheBinaries caches Kubernetes binaries in the foreground
func doCacheBinaries(k8sVersion string) error {
	return machine.CacheBinariesForBootstrapper(k8sVersion, viper.GetString(cmdcfg.Bootstrapper))
}

// beginDownloadKicBaseImage downloads the kic image
func beginDownloadKicBaseImage(g *errgroup.Group, cc *config.ClusterConfig, downloadOnly bool) {
	if cc.Driver != "docker" {
		// TODO: driver == "podman"
		klog.Info("Driver isn't docker, skipping base image download")
		return
	}
	if image.ExistsImageInDaemon(cc.KicBaseImage) {
		klog.Infof("%s exists in daemon, skipping pull", cc.KicBaseImage)
		return
	}

	klog.Infof("Beginning downloading kic base image for %s with %s", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
	register.Reg.SetStep(register.PullingBaseImage)
	out.T(style.Pulling, "Pulling base image ...")
	g.Go(func() error {
		baseImg := cc.KicBaseImage
		if baseImg == kic.BaseImage && len(cc.KubernetesConfig.ImageRepository) != 0 {
			baseImg = strings.Replace(baseImg, "gcr.io/k8s-minikube", cc.KubernetesConfig.ImageRepository, 1)
		}
		var finalImg string
		// If we end up using a fallback image, notify the user
		defer func() {
			if finalImg != "" && finalImg != baseImg {
				out.WarningT(fmt.Sprintf("minikube was unable to download %s, but successfully downloaded %s as a fallback image", image.Tag(baseImg), image.Tag(finalImg)))
				cc.KicBaseImage = finalImg
			}
		}()
		for _, img := range append([]string{baseImg}, kic.FallbackImages...) {
			if err := image.LoadFromTarball(driver.Docker, img); err == nil {
				klog.Infof("successfully loaded %s from cached tarball", img)
				// strip the digest from the img before saving it in the config
				// because loading an image from tarball to daemon doesn't load the digest
				finalImg = img
				return nil
			}
			klog.Infof("Downloading %s to local daemon", img)
			err := image.WriteImageToDaemon(img)
			if err == nil {
				klog.Infof("successfully downloaded %s", img)
				finalImg = img
				return nil
			}
			if downloadOnly {
				if err := image.SaveToDir([]string{img}, constants.ImageCacheDir); err == nil {
					klog.Infof("successfully saved %s as a tarball", img)
					finalImg = img
					return nil
				}
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
				out.ErrT(style.Connectivity, "Unfortunately, could not download the base image {{.image_name}} ", out.V{"image_name": strings.Split(kic.BaseImage, "@")[0]})
				out.WarningT("In order to use the fall back image, you need to log in to the github packages registry")
				out.T(style.Documentation, `Please visit the following link for documentation around this: 
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

// WaitCacheRequiredImages blocks until the required images are all cached.
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
	return image.SaveToDir(images, constants.ImageCacheDir)
}

// CacheAndLoadImagesInConfig loads the images currently in the config file
// called by 'start' and 'cache reload' commands.
func CacheAndLoadImagesInConfig() error {
	images, err := imagesInConfigFile()
	if err != nil {
		return errors.Wrap(err, "images")
	}
	if len(images) == 0 {
		return nil
	}
	return machine.CacheAndLoadImages(images)
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
