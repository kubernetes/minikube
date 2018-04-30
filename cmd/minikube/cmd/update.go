/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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
	"k8s.io/minikube/pkg/minikube/update"
	"k8s.io/minikube/pkg/version"
	"os"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates Minikube to the latest version",
	Long:  `Updates Minikube to the latest version`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Explicitly disable update checking for the version command
		enableUpdateNotification = false
	},
	Run: func(command *cobra.Command, args []string) {
		currentVersion, err := version.GetSemverVersion()
		if err != nil {
			fmt.Fprint(os.Stderr, "Error fetching current version")
			os.Exit(1)
		}

		latestVersion, err := update.LatestVersion()
		if err != nil {
			fmt.Fprint(os.Stderr, "Error fetching latest version from internet")
			os.Exit(1)
		}

		if update.IsNewerVersion(currentVersion, latestVersion) {
			if err := update.Update(latestVersion); err != nil {
				fmt.Fprintf(os.Stderr, "Error updating to latest version: %s", err)
				os.Exit(1)
			}

			fmt.Printf("Updated successfully to version %s%s\n", version.VersionPrefix, latestVersion)
		}
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
