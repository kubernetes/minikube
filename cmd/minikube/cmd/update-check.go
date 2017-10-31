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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/version"
)

var updateCheckCmd = &cobra.Command{
	Use:   "update-check",
	Short: "Print current and latest version number",
	Long:  `Print current and latest version number`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Explicitly disable update checking for the version command
		enableUpdateNotification = false
	},
	Run: func(command *cobra.Command, args []string) {
		url := constants.GithubMinikubeReleasesURL
		r, err := notify.GetAllVersionsFromURL(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching latest version from internet")
			os.Exit(1)
		}

		if len(r) < 1 {
			fmt.Fprintf(os.Stderr, "Got empty version list from server")
			os.Exit(2)
		}

		fmt.Println("CurrentVersion:", version.GetVersion())
		fmt.Println("LatestVersion:", r[0].Name)
	},
}

func init() {
	RootCmd.AddCommand(updateCheckCmd)
}
