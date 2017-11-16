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

		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		machine.CacheImages(args, constants.ImageCacheDir)

		h, err := api.Load(config.GetMachineName())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting api client: %s\n", err)
			os.Exit(1)
		}

		client, err := sshutil.NewSSHClient(h.Driver)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting ssh client: %s\n", err)
			os.Exit(1)
		}
		cmdRunner := bootstrapper.NewSSHRunner(client)

		machine.LoadImages(cmdRunner, args, constants.ImageCacheDir)
		err = cmdConfig.AddToConfigArray("cache", images)
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
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		h, err := api.Load(config.GetMachineName())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting api client: %s\n", err)
			os.Exit(1)
		}

		client, err := sshutil.NewSSHClient(h.Driver)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting ssh client: %s\n", err)
			os.Exit(1)
		}
		cmdRunner := bootstrapper.NewSSHRunner(client)
		machine.DeleteImages(cmdRunner, args)

		err = cmdConfig.DeleteFromConfigArray("cache", args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting images from cache: %s\n", err)
			os.Exit(1)
		}

	},
}

func init() {
	cacheCmd.AddCommand(addCacheCmd)
	cacheCmd.AddCommand(deleteCacheCmd)
	RootCmd.AddCommand(cacheCmd)
}
