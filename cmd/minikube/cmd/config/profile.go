/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	pkgConfig "k8s.io/minikube/pkg/minikube/config"
)

var ProfileCmd = &cobra.Command{
	Use:   "profile MINIKUBE_PROFILE_NAME.  You can return to the default minikube profile by running `minikube profile default`",
	Short: "Profile sets the current minikube profile",
	Long:  "profile sets the current minikube profile.  This is used to run and manage multiple minikube instance.  You can return to the default minikube profile by running `minikube profile default`",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: minikube profile MINIKUBE_PROFILE_NAME")
			os.Exit(1)
		}

		profile := args[0]
		if profile == "default" {
			profile = "minikube"
		}
		err := Set(pkgConfig.MachineProfile, profile)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
		} else {
			fmt.Fprintln(os.Stdout, fmt.Sprintf("minikube profile was successfully set to %s", profile))
		}
	},
}
