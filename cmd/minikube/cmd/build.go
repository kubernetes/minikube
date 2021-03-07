/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a container image",
	Long: `Build a container image, using the container runtime.
Examples:
minikube build .`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			exit.Message(reason.Usage, "minikube build -- [OPTIONS] PATH | URL | -")
		}
		// Cache and load images into docker daemon
		profile, err := config.LoadProfile(viper.GetString(config.ProfileName))
		if err != nil {
			exit.Error(reason.Usage, "loading profile", err)
		}
		img := args[0]
		if err := machine.BuildImage(img, "test:latest", []*config.Profile{profile}); err != nil {
			exit.Error(reason.GuestImageBuild, "Failed to build image", err)
		}
	},
}

func init() {
}
