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
	"os"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/exit"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/minikube/constants"
)

// ProfileCmd represents the profile command
var ProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "profiles gets the list of all the present profiles",
	Long:  "profiles displays a list of all the profiles which have been created earlier.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			exit.Usage("usage: minikube profile [MINIKUBE_PROFILE_NAME]")
		}
		profiles, err := pkgutil.GetProfiles(constants.KubeconfigPath)
		if err != nil {
			exit.WithError("Failed to fetch profiles", err)
		}

		// check length og profiles, if zero then print suitable message
		if len(profiles) == 0 {
			console.OutLn("No profiles created yet")
			os.Exit(0)
		}

		for _, profile := range profiles {
			console.OutLn(profile)
		}
	},
}
