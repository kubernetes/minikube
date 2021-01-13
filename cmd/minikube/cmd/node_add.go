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
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util"
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
		config.TagPrimaryControlPlane(cc)

		if driver.BareMetal(cc.Driver) {
			out.FailureT("none driver does not support multi-node clusters")
		}

		name := node.Name(len(cc.Nodes) + 1)

		if cp {
			out.Step(style.Happy, "Adding control plane node {{.name}} to cluster {{.cluster}}", out.V{"name": name, "cluster": cc.Name})
		} else {
			out.Step(style.Happy, "Adding node {{.name}} to cluster {{.cluster}}", out.V{"name": name, "cluster": cc.Name})
		}

		// TODO: Deal with parameters better. Ideally we should be able to acceot any node-specific minikube start params here.
		n := config.Node{
			Name:                name,
			Worker:              worker,
			ControlPlane:        cp,
			PrimaryControlPlane: false,
			KubernetesVersion:   cc.KubernetesConfig.KubernetesVersion,
		}

		if n.ControlPlane {
			err := util.CheckMultiControlPlaneVersion(cc.KubernetesConfig.KubernetesVersion)
			if err != nil {
				exit.Error(reason.KubernetesTooOld, "target kubernetes version too old", err)
			}
			n.Port = cc.KubernetesConfig.NodePort
		}

		// Make sure to decrease the default amount of memory we use per VM if this is the first worker node
		if len(cc.Nodes) == 1 {
			if viper.GetString(memory) == "" {
				cc.Memory = 2200
			}

			if !cc.MultiNodeRequested || cni.IsDisabled(*cc) {
				warnAboutMultiNodeCNI()
			}
		}

		register.Reg.SetStep(register.InitialSetup)
		if err := node.Add(cc, n, false); err != nil {
			_, err := maybeDeleteAndRetry(cmd, *cc, n, nil, err)
			if err != nil {
				exit.Error(reason.GuestNodeAdd, "failed to add node", err)
			}
		}

		if err := config.SaveProfile(cc.Name, cc); err != nil {
			exit.Error(reason.HostSaveProfile, "failed to save config", err)
		}

		out.Step(style.Ready, "Successfully added {{.name}} to {{.cluster}}!", out.V{"name": name, "cluster": cc.Name})
	},
}

func init() {
	// TODO(https://github.com/kubernetes/minikube/issues/7366): We should figure out which minikube start flags to actually import
	nodeAddCmd.Flags().BoolVar(&cp, "control-plane", false, "If true, the node added will also be a control plane in addition to a worker.")
	nodeAddCmd.Flags().BoolVar(&worker, "worker", true, "If true, the added node will be marked for work. Defaults to true.")
	nodeAddCmd.Flags().Bool(deleteOnFailure, false, "If set, delete the current cluster if start fails and try again. Defaults to false.")

	nodeCmd.AddCommand(nodeAddCmd)
}
