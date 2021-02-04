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
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/reason"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Load a local image into minikube",
	Long:  "Load a local image into minikube",
}

// loadImageCmd represents the image load command
var loadImageCmd = &cobra.Command{
	Use:   "load",
	Short: "Load a local image into minikube",
	Long:  "Load a local image into minikube",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exit.Message(reason.Usage, "Please provide an image in your local daemon or a path to an image tarball to load into minikube via <minikube image load IMAGE_NAME>")
		}
		// Cache and load images into docker daemon
		profile := viper.GetString(config.ProfileName)
		img := args[0]
		if err := machine.LoadImage(profile, img); err != nil {
			exit.Error(reason.GuestImageLoad, "Failed to load image", err)
		}
	},
}

func init() {
	imageCmd.AddCommand(loadImageCmd)
}
