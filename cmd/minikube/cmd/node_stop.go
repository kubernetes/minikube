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
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var nodeStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops a node in a cluster.",
	Long:  "Stops a node in a cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exit.Message(reason.Usage, "Usage: minikube node stop [name]")
		}

		name := args[0]
		api, cc := mustload.Partial(ClusterFlagValue())

		n, _, err := node.Retrieve(*cc, name)
		if err != nil {
			exit.Error(reason.GuestNodeRetrieve, "retrieving node", err)
		}

		machineName := config.MachineName(*cc, *n)
		node.MustReset(*cc, *n, api, machineName)
		err = machine.StopHost(api, machineName)
		if err != nil {
			out.FatalT("Failed to stop node {{.name}}", out.V{"name": name})
		}
		out.Step(style.Stopped, "Successfully stopped node {{.name}}", out.V{"name": machineName})
	},
}

func init() {
	nodeCmd.AddCommand(nodeStopCmd)
}
