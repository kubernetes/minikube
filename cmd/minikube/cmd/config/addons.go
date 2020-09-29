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
	"github.com/spf13/cobra"

	"k8s.io/klog/v2"
)

// AddonsCmd represents the addons command
var AddonsCmd = &cobra.Command{
	Use:   "addons SUBCOMMAND [flags]",
	Short: "Enable or disable a minikube addon",
	Long:  `addons modifies minikube addons files using subcommands like "minikube addons enable dashboard"`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			klog.Errorf("help: %v", err)
		}
	},
}
