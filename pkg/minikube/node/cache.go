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
	"os"
	"runtime"

	"github.com/golang/glog"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/image"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

// beginCacheRequiredImages caches images required for kubernetes version in the background
func beginCacheRequiredImages(g *errgroup.Group, imageRepository string, k8sVersion string) {
	if !viper.GetBool("cache-images") {
		return
	}

	g.Go(func() error {
		return machine.CacheImagesForBootstrapper(imageRepository, k8sVersion, viper.GetString(cmdcfg.Bootstrapper))
	})
}

func handleDownloadOnly(cacheGroup *errgroup.Group, k8sVersion string) {
	// If --download-only, complete the remaining downloads and exit.
	if !viper.GetBool("download-only") {
		return
	}
	if err := doCacheBinaries(k8sVersion); err != nil {
		exit.WithError("Failed to cache binaries", err)
	}
	if _, err := CacheKubectlBinary(k8sVersion); err != nil {
		exit.WithError("Failed to cache kubectl", err)
	}
	waitCacheRequiredImages(cacheGroup)
	if err := saveImagesToTarFromConfig(); err != nil {
		exit.WithError("Failed to cache images to tar", err)
	}
	out.T(out.Check, "Download complete!")
	os.Exit(0)

}

// CacheKubectlBinary caches the kubectl binary
func CacheKubectlBinary(k8sVerison string) (string, error) {
	binary := "kubectl"
	if runtime.GOOS == "windows" {
		binary = "kubectl.exe"
	}

	return machine.CacheBinary(binary, k8sVerison, runtime.GOOS, runtime.GOARCH)
}

// doCacheBinaries caches Kubernetes binaries in the foreground
func doCacheBinaries(k8sVersion string) error {
	return machine.CacheBinariesForBootstrapper(k8sVersion, viper.GetString(cmdcfg.Bootstrapper))
}

// waitCacheRequiredImages blocks until the required images are all cached.
func waitCacheRequiredImages(g *errgroup.Group) {
	if !viper.GetBool(cacheImages) {
		return
	}
	if err := g.Wait(); err != nil {
		glog.Errorln("Error caching images: ", err)
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

func imagesInConfigFile() ([]string, error) {
	configFile, err := config.ReadConfig(localpath.ConfigFile)
	if err != nil {
		return nil, err
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

// CacheAndLoadImagesInConfig loads the images currently in the config file
// called by 'start' and 'cache reload' commands.
func CacheAndLoadImagesInConfig() error {
	images, err := imagesInConfigFile()
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return nil
	}
	return machine.CacheAndLoadImages(images)
}
