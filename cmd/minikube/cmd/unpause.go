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
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
)

// unpauseCmd represents the docker-pause command
var unpauseCmd = &cobra.Command{
	Use:   "unpause",
	Short: "unpause Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		cname := ClusterFlagValue()
		register.SetEventLogPath(localpath.EventLog(cname))

		co := mustload.Running(cname)
		register.Reg.SetStep(register.Unpausing)

		glog.Infof("namespaces: %v keys: %v", namespaces, viper.AllSettings())
		if allNamespaces {
			namespaces = nil //all
		} else {
			if len(namespaces) == 0 {
				exit.WithCodeT(exit.ProgramUsage, "Use -A to specify all namespaces")
			}
		}

		ids := []string{}

		for _, n := range co.Config.Nodes {
			glog.Infof("node: %+v", n)

			// Use node-name if available, falling back to cluster name
			name := n.Name
			if n.Name == "" {
				name = co.Config.Name
			}

			out.T(out.Pause, "Unpausing node {{.name}} ... ", out.V{"name": name})

			machineName := driver.MachineName(*co.Config, n)
			host, err := machine.LoadHost(co.API, machineName)
			if err != nil {
				exit.WithError(exit.ProgramError, "Error getting host", err)
			}

			r, err := machine.CommandRunner(host)
			if err != nil {
				exit.WithError(exit.ProgramError, "Failed to get command runner", err)
			}

			cr, err := cruntime.New(cruntime.Config{Type: co.Config.KubernetesConfig.ContainerRuntime, Runner: r})
			if err != nil {
				exit.WithError(exit.ProgramError, "Failed runtime", err)
			}

			uids, err := cluster.Unpause(cr, r, namespaces)
			if err != nil {
				exit.WithError(exit.ProgramError, "Pause", err)
			}
			ids = append(ids, uids...)
		}

		register.Reg.SetStep(register.Done)

		if namespaces == nil {
			out.T(out.Pause, "Unpaused {{.count}} containers", out.V{"count": len(ids)})
		} else {
			out.T(out.Pause, "Unpaused {{.count}} containers in: {{.namespaces}}", out.V{"count": len(ids), "namespaces": strings.Join(namespaces, ", ")})
		}
	},
}

func init() {
	unpauseCmd.Flags().StringSliceVarP(&namespaces, "--namespaces", "n", constants.DefaultNamespaces, "namespaces to unpause")
	unpauseCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "If set, unpause all namespaces")
}
