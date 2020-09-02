/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/reason"
)

// nodeCmd represents the set of node subcommands
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Add, remove, or list additional nodes",
	Long:  "Operations on nodes",
	Run: func(cmd *cobra.Command, args []string) {
		exit.Message(reason.Usage, "Usage: minikube node [add|start|stop|delete|list]")
	},
}
