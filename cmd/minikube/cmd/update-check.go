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
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
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
			exit.WithError("Unable to fetch latest version info", err)
		}

		if len(r) < 1 {
			exit.WithCode(exit.Data, "Update server returned an empty list")
		}

		console.OutLn("CurrentVersion: %s", version.GetVersion())
		console.OutLn("LatestVersion: %s", r[0].Name)
	},
}

func init() {
	RootCmd.AddCommand(updateCheckCmd)
}
