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
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

// ipCmd represents the ip command
var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Retrieves the IP address of the specified node",
	Long:  `Retrieves the IP address of the specified node, and writes it to STDOUT.`,
	Run: func(cmd *cobra.Command, args []string) {
		co := mustload.Running(ClusterFlagValue())
		n, _, err := node.Retrieve(*co.Config, nodeName)
		if err != nil {
			exit.Error(reason.GuestNodeRetrieve, "retrieving node", err)
		}

		out.Ln(n.IP)
	},
}

func init() {
	ipCmd.Flags().StringVarP(&nodeName, "node", "n", "", "The node to get IP. Defaults to the primary control plane.")
}
