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
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var addonsDisableCmd = &cobra.Command{
	Use:   "disable ADDON_NAME",
	Short: "Disables the addon w/ADDON_NAME within minikube (example: minikube addons disable dashboard). For a list of available addons use: minikube addons list ",
	Long:  "Disables the addon w/ADDON_NAME within minikube (example: minikube addons disable dashboard). For a list of available addons use: minikube addons list ",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "usage: minikube addons disable ADDON_NAME")
		}

		addon := args[0]
		if addon == "heapster" {
			exit.Message(reason.AddonUnsupported, "The heapster addon is depreciated. please try to disable metrics-server instead")
		}
		err := addons.SetAndSave(ClusterFlagValue(), addon, "false")
		if err != nil {
			exit.Error(reason.InternalDisable, "disable failed", err)
		}
		out.Step(style.AddonDisable, `"The '{{.minikube_addon}}' addon is disabled`, false, out.V{"minikube_addon": addon})
	},
}

func init() {
	AddonsCmd.AddCommand(addonsDisableCmd)
}
