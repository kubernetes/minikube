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
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

var nodeStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops a node in a cluster.",
	Long:  "Stops a node in a cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exit.UsageT("Usage: minikube node stop [name]")
		}

		name := args[0]

		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("creating api client", err)
		}

		err = machine.StopHost(api, name)
		if err != nil {
			out.FatalT("Failed to stop node {{.name}}", out.V{"name": name})
		}
	},
}

func init() {
	nodeStopCmd.Flags().String("name", "", "The name of the node to delete")
	nodeCmd.AddCommand(nodeStopCmd)
}
