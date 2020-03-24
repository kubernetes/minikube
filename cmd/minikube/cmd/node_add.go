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
	"github.com/spf13/pflag"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
)

var (
	cp     bool
	worker bool
)
var nodeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a node to the given cluster.",
	Long:  "Adds a node to the given cluster config, and starts it.",
	Run: func(cmd *cobra.Command, args []string) {
		co := mustload.Healthy(ClusterFlagValue())
		cc := co.Config

		if driver.BareMetal(cc.Driver) {
			out.ErrT(out.FailureType, "none driver does not support multi-node clusters")
		}

		name := node.Name(len(cc.Nodes) + 1)

		out.T(out.Happy, "Adding node {{.name}} to cluster {{.cluster}}", out.V{"name": name, "cluster": cc.Name})

		// TODO: Deal with parameters better. Ideally we should be able to acceot any node-specific minikube start params here.
		n := config.Node{
			Name:              name,
			Worker:            worker,
			ControlPlane:      cp,
			KubernetesVersion: cc.KubernetesConfig.KubernetesVersion,
		}

		if err := node.Add(cc, n); err != nil {
			exit.WithError("Error adding node to cluster", err)
		}

		out.T(out.Ready, "Successfully added {{.name}} to {{.cluster}}!", out.V{"name": name, "cluster": cc.Name})
	},
}

func init() {
	nodeAddCmd.Flags().BoolVar(&cp, "control-plane", false, "If true, the node added will also be a control plane in addition to a worker.")
	nodeAddCmd.Flags().BoolVar(&worker, "worker", true, "If true, the added node will be marked for work. Defaults to true.")
	//We should figure out which of these flags to actually import
	startCmd.Flags().Visit(
		func(f *pflag.Flag) {
			nodeAddCmd.Flags().AddFlag(f)
		},
	)
	nodeCmd.AddCommand(nodeAddCmd)
}
