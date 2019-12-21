/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"github.com/spf13/cobra"
	cmdConfig "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/image"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
)

// cacheImageConfigKey is the config field name used to store which images we have previously cached
const cacheImageConfigKey = "cache"

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Add or delete an image from the local cache.",
	Long:  "Add or delete an image from the local cache.",
}

// addCacheCmd represents the cache add command
var addCacheCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an image to local cache.",
	Long:  "Add an image to local cache.",
	Run: func(cmd *cobra.Command, args []string) {
		// Cache and load images into docker daemon
		if err := machine.CacheAndLoadImages(args); err != nil {
			exit.WithError("Failed to cache and load images", err)
		}
		// Add images to config file
		if err := cmdConfig.AddToConfigMap(cacheImageConfigKey, args); err != nil {
			exit.WithError("Failed to update config", err)
		}
	},
}

// deleteCacheCmd represents the cache delete command
var deleteCacheCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an image from the local cache.",
	Long:  "Delete an image from the local cache.",
	Run: func(cmd *cobra.Command, args []string) {
		// Delete images from config file
		if err := cmdConfig.DeleteFromConfigMap(cacheImageConfigKey, args); err != nil {
			exit.WithError("Failed to delete images from config", err)
		}
		// Delete images from cache/images directory
		if err := image.DeleteFromCacheDir(args); err != nil {
			exit.WithError("Failed to delete images", err)
		}
	},
}

// reloadCacheCmd represents the cache reload command
var reloadCacheCmd = &cobra.Command{
	Use:   "reload",
	Short: "reload cached images.",
	Long:  "reloads images previously added using the 'cache add' subcommand",
	Run: func(cmd *cobra.Command, args []string) {
		err := cacheAndLoadImagesInConfig()
		if err != nil {
			exit.WithError("Failed to reload cached images", err)
		}
	},
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

// cacheAndLoadImagesInConfig loads the images currently in the config file
// called by 'start' and 'cache reload' commands.
func cacheAndLoadImagesInConfig() error {
	images, err := imagesInConfigFile()
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return nil
	}
	return machine.CacheAndLoadImages(images)
}

func init() {
	cacheCmd.AddCommand(addCacheCmd)
	cacheCmd.AddCommand(deleteCacheCmd)
	cacheCmd.AddCommand(reloadCacheCmd)
}
