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
	"os"

	"github.com/spf13/cobra"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
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

		config, err := cfg.Load()
		if err != nil && !os.IsNotExist(err) {
			exit.WithCodeT(exit.Data, "Unable to load config: {{.error}}", out.V{"error": err})
		}

		config.CachedImages = append(config.CachedImages, args...)
		if err = cfg.CreateProfile(config.Name, config); err != nil {
			exit.WithCodeT(exit.Data, "Unable to save config: {{.error}}", out.V{"error": err})
		}

	},
}

// deleteCacheCmd represents the cache delete command
var deleteCacheCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an image from the local cache.",
	Long:  "Delete an image from the local cache.",
	Run: func(cmd *cobra.Command, args []string) {
		// Delete images from cache/images directory
		if err := machine.DeleteFromImageCacheDir(args); err != nil {
			exit.WithError("Failed to delete images", err)
		}

		config, err := cfg.Load()
		if err != nil && !os.IsNotExist(err) {
			exit.WithCodeT(exit.Data, "Unable to load config: {{.error}}", out.V{"error": err})
		}

		updatedList := []string{}
		for _, img := range config.CachedImages {
			toAdd := true
			for _, toDel := range args {
				if img == toDel {
					toAdd = false
				}
			}
			if toAdd {
				updatedList = append(updatedList, img)
			}
		}
		config.CachedImages = updatedList

		if err = cfg.CreateProfile(config.Name, config); err != nil {
			exit.WithError("Failed to delete images from config", err)
		}

	},
}

func init() {
	cacheCmd.AddCommand(addCacheCmd)
	cacheCmd.AddCommand(deleteCacheCmd)
}
