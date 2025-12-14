/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/reason"
)

var dir string

// licenseCmd represents the credits command
var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Outputs the licenses of dependencies to a directory",
	Long:  "Outputs the licenses of dependencies to a directory",
	Run: func(_ *cobra.Command, _ []string) {
		if err := download.Licenses(dir); err != nil {
			exit.Error(reason.InetLicenses, "Failed to download licenses", err)
		}
	},
}

func init() {
	licenseCmd.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to output licenses to")
}
