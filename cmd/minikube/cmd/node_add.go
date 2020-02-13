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
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
)

var (
	nodeName string
	cp       bool
	worker   bool
)
var nodeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a node to the given cluster.",
	Long:  "Adds a node to the given cluster config, and starts it.",
	Run: func(cmd *cobra.Command, args []string) {
		profile := viper.GetString(config.MachineProfile)
		mc, err := config.Load(profile)
		if err != nil {
			exit.WithError("Error getting config", err)
		}
		name := nodeName
		if nodeName == "" {
			name = profile + strconv.Itoa(len(mc.Nodes)+1)
		}
		out.T(out.Happy, "Adding node {{.name}} to cluster {{.cluster}}", out.V{"name": name, "cluster": profile})

		n, err := node.Add(mc, name, cp, worker, "", profile)
		if err != nil {
			exit.WithError("Error adding node to cluster", err)
		}

		_, err = node.Start(*mc, *n, false, nil)
		if err != nil {
			exit.WithError("Error starting node", err)
		}

		out.T(out.Ready, "Successfully added {{.name}} to {{.cluster}}!", out.V{"name": name, "cluster": profile})
	},
}

func init() {
	nodeAddCmd.Flags().StringVar(&nodeName, "name", "", "The name of the node to add.")
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
