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
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var nodeStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a node.",
	Long:  "Starts an existing stopped node in a cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exit.Message(reason.Usage, "Usage: minikube node start [name]")
		}

		api, cc := mustload.Partial(ClusterFlagValue())
		name := args[0]

		n, _, err := node.Retrieve(*cc, name)
		if err != nil {
			exit.Error(reason.GuestNodeRetrieve, "retrieving node", err)
		}

		machineName := config.MachineName(*cc, *n)
		if machine.IsRunning(api, machineName) {
			out.Step(style.Check, "{{.name}} is already running", out.V{"name": name})
			os.Exit(0)
		}

		register.Reg.SetStep(register.InitialSetup)
		r, p, m, h, err := node.Provision(cc, n, viper.GetBool(deleteOnFailure))
		if err != nil {
			exit.Error(reason.GuestNodeProvision, "provisioning host for node", err)
		}

		s := node.Starter{
			Runner:         r,
			PreExists:      p,
			MachineAPI:     m,
			Host:           h,
			Cfg:            cc,
			Node:           n,
			ExistingAddons: nil,
		}

		_, err = node.Start(s)
		if err != nil {
			_, err := maybeDeleteAndRetry(cmd, *cc, *n, nil, err)
			if err != nil {
				node.ExitIfFatal(err)
				exit.Error(reason.GuestNodeStart, "failed to start node", err)
			}
		}
		out.Step(style.Happy, "Successfully started node {{.name}}!", out.V{"name": machineName})
	},
}

func init() {
	nodeStartCmd.Flags().Bool(deleteOnFailure, false, "If set, delete the current cluster if start fails and try again. Defaults to false.")
	nodeCmd.AddCommand(nodeStartCmd)
}
