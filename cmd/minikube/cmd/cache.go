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
	"fmt"
	"github.com/spf13/cobra"
	cmdConfig "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"os"
)

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
		err := cacheAndLoadImages(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error caching and loading images: %s\n", err)
			os.Exit(1)
		}
		// Add images to config file
		err = cmdConfig.AddToConfigArray("cache", args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding cached images to config file: %s\n", err)
			os.Exit(1)
		}
	},
}

// deleteCacheCmd represents the cache delete command
var deleteCacheCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an image from the local cache.",
	Long:  "Delete an image from the local cache.",
	Run: func(cmd *cobra.Command, args []string) {

		cmdRunner, err := getCommandRunner()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting command runner: %s\n", err)
			os.Exit(1)
		}
		// Delete images from docker daemon
		err = machine.DeleteImages(cmdRunner, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting images: %s\n", err)
			os.Exit(1)
		}
		// Delete images from config file
		err = cmdConfig.DeleteFromConfigArray("cache", args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting images from config file: %s\n", err)
			os.Exit(1)
		}

	},
}

// LoadCachedImagesInConfigFile loads the images currently in the config file (minikube start)
func LoadCachedImagesInConfigFile() error {
	configFile, err := config.ReadConfig()
	if err != nil {
		return err
	}

	values := configFile["cache"]

	if values == nil {
		return nil
	}

	var images []string

	for _, v := range values.([]interface{}) {
		images = append(images, v.(string))
	}

	return cacheAndLoadImages(images)

}

func cacheAndLoadImages(images []string) error {

	err := machine.CacheImages(images, constants.ImageCacheDir)
	if err != nil {
		return err
	}

	cmdRunner, err := getCommandRunner()
	if err != nil {
		return err
	}

	return machine.LoadImages(cmdRunner, images, constants.ImageCacheDir)

}

func getCommandRunner() (*bootstrapper.SSHRunner, error) {
	api, err := machine.NewAPIClient()
	if err != nil {
		return nil, err
	}
	defer api.Close()
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return nil, err
	}

	client, err := sshutil.NewSSHClient(h.Driver)
	if err != nil {
		return nil, err
	}
	return bootstrapper.NewSSHRunner(client), nil

}

func init() {
	cacheCmd.AddCommand(addCacheCmd)
	cacheCmd.AddCommand(deleteCacheCmd)
	RootCmd.AddCommand(cacheCmd)
}
