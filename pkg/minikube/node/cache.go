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
	"path"
	"runtime"
	"strings"

	"k8s.io/minikube/pkg/minikube/detect"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
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
func beginCacheKubernetesImages(g *errgroup.Group, imageRepository string, k8sVersion string, cRuntime string, driverName string) {
	// TODO: remove imageRepository check once #7695 is fixed
	if imageRepository == "" && download.PreloadExists(k8sVersion, cRuntime, driverName) {
		klog.Info("Caching tarball of preloaded images")
		err := download.Preload(k8sVersion, cRuntime, driverName)
		if err == nil {
			klog.Infof("Finished verifying existence of preloaded tar for %s on %s", k8sVersion, cRuntime)
			return // don't cache individual images if preload is successful.
		}
		klog.Warningf("Error downloading preloaded artifacts will continue without preload: %v", err)
	}

	if !viper.GetBool(cacheImages) {
		return
	}

	g.Go(func() error {
		return machine.CacheImagesForBootstrapper(imageRepository, k8sVersion)
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
	if driver.IsKIC(driverName) {
		waitDownloadKicBaseImage(kicGroup)
	}
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

	return download.Binary(binary, k8sVersion, runtime.GOOS, runtime.GOARCH, binaryURL)
}

// doCacheBinaries caches Kubernetes binaries in the foreground
func doCacheBinaries(k8sVersion, containerRuntime, driverName, binariesURL string) error {
	existingBinaries := constants.KubernetesReleaseBinaries
	if !download.PreloadExists(k8sVersion, containerRuntime, driverName) {
		existingBinaries = nil
	}
	return machine.CacheBinariesForBootstrapper(k8sVersion, existingBinaries, binariesURL)
}

// beginDownloadKicBaseImage downloads the kic image
func beginDownloadKicBaseImage(g *errgroup.Group, cc *config.ClusterConfig, downloadOnly bool) {

	klog.Infof("Beginning downloading kic base image for %s with %s", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
	register.Reg.SetStep(register.PullingBaseImage)
	out.Step(style.Pulling, "Pulling base image {{.kicVersion}} ...", out.V{"kicVersion": kic.Version})
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
					out.WarningT(fmt.Sprintf("minikube was unable to download %s, but successfully downloaded %s as a fallback image", image.Tag(baseImg), finalImg))
				}
			}
		}()
		// first we try to download the kicbase image (and fall back images) from docker registry
		var err error
		for _, img := range append([]string{baseImg}, kic.FallbackImages...) {

			if driver.IsDocker(cc.Driver) && download.ImageExistsInDaemon(img) && !downloadOnly {
				klog.Infof("%s exists in daemon, skipping load", img)
				finalImg = img
				return nil
			}

			klog.Infof("Downloading %s to local cache", img)
			err = download.ImageToCache(img)
			if err == nil {
				klog.Infof("successfully saved %s as a tarball", img)
			}
			if downloadOnly && err == nil {
				return nil
			}

			if cc.Driver == driver.Podman {
				return fmt.Errorf("not yet implemented, see issue #8426")
			}
			if driver.IsDocker(cc.Driver) && err == nil {
				klog.Infof("Loading %s from local cache", img)
				if finalImg, err = download.CacheToDaemon(img); err == nil {
					klog.Infof("successfully loaded and using %s from cached tarball", img)
					return nil
				}
			}
			klog.Infof("failed to download %s, will try fallback image if available: %v", img, err)
		}
		// second if we failed to download any fallback image
		// that means probably all registries are blocked by network issues
		// we can try to download the image from minikube release page

		// if we reach here, that means the user cannot have access to any docker registry
		// this means the user is very likely to have a network issue
		// downloading from github via http is the last resort, and we should remind the user
		// that he should at least get access to github
		// print essential warnings
		out.WarningT("minikube cannot pull kicbase image from any docker registry, and is trying to download kicbase tarball from github release page via HTTP.")
		out.WarningT("It's very likely that you have an internet issue. Please ensure that you can access the internet at least via HTTP, directly or with proxy. Currently your proxy configure is:")
		envs := []string{"HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy", "ALL_PROXY", "NO_PROXY"}
		for _, env := range envs {
			if v := os.Getenv(env); v != "" {
				out.Infof("{{.env}}={{.value}}", out.V{"env": env, "value": v})
			}
		}
		out.Ln("")

		kicbaseVersion := strings.Split(kic.Version, "-")[0]
		_, err = download.GHKicbaseTarballToCache(kicbaseVersion)
		if err != nil {
			klog.Infof("failed to download kicbase from github")
			return fmt.Errorf("failed to download kic base image or any fallback image")
		}

		klog.Infof("successfully downloaded kicbase as fall back image from github")
		if !downloadOnly && driver.IsDocker(cc.Driver) {
			if finalImg, err = download.CacheToDaemon(fmt.Sprintf("kicbase/stable:%s", kicbaseVersion)); err == nil {
				klog.Infof("successfully loaded and using kicbase from tarball on github")
			} else {
				return fmt.Errorf("failed to load kic base image into docker: %v", err)
			}
		}
		return nil

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
