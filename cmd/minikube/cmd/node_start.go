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
	"os"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
)

var nodeStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a node.",
	Long:  "Starts an existing stopped node in a cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exit.UsageT("Usage: minikube node start [name]")
		}

		api, cc := mustload.Partial(ClusterFlagValue())
		name := args[0]

		if machine.IsRunning(api, name) {
			out.T(out.Check, "{{.name}} is already running", out.V{"name": name})
			os.Exit(0)
		}

		n, _, err := node.Retrieve(cc, name)
		if err != nil {
			exit.WithError("retrieving node", err)
		}

		// Start it up baby
		node.Start(*cc, *n, nil, false)
	},
}

func init() {
	nodeStartCmd.Flags().String("name", "", "The name of the node to start")
	nodeCmd.AddCommand(nodeStartCmd)
}
