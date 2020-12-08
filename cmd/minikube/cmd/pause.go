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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/klog/v2"
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
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var (
	namespaces    []string
	allNamespaces bool
)

// pauseCmd represents the docker-pause command
var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "pause Kubernetes",
	Run:   runPause,
}

func runPause(cmd *cobra.Command, args []string) {
	out.SetJSON(outputFormat == "json")
	co := mustload.Running(ClusterFlagValue())
	register.SetEventLogPath(localpath.EventLog(ClusterFlagValue()))
	register.Reg.SetStep(register.Pausing)

	klog.Infof("namespaces: %v keys: %v", namespaces, viper.AllSettings())
	if allNamespaces {
		namespaces = nil // all
	} else if len(namespaces) == 0 {
		exit.Message(reason.Usage, "Use -A to specify all namespaces")
	}

	ids := []string{}

	for _, n := range co.Config.Nodes {
		// Use node-name if available, falling back to cluster name
		name := n.Name
		if n.Name == "" {
			name = co.Config.Name
		}

		out.Step(style.Pause, false, "Pausing node {{.name}} ... ", out.V{"name": name})

		host, err := machine.LoadHost(co.API, driver.MachineName(*co.Config, n))
		if err != nil {
			exit.Error(reason.GuestLoadHost, "Error getting host", err)
		}

		r, err := machine.CommandRunner(host)
		if err != nil {
			exit.Error(reason.InternalCommandRunner, "Failed to get command runner", err)
		}

		cr, err := cruntime.New(cruntime.Config{Type: co.Config.KubernetesConfig.ContainerRuntime, Runner: r})
		if err != nil {
			exit.Error(reason.InternalNewRuntime, "Failed runtime", err)
		}

		uids, err := cluster.Pause(cr, r, namespaces)
		if err != nil {
			exit.Error(reason.GuestPause, "Pause", err)
		}
		ids = append(ids, uids...)
	}

	register.Reg.SetStep(register.Done)
	if namespaces == nil {
		out.Step(style.Unpause, false, "Paused {{.count}} containers", out.V{"count": len(ids)})
	} else {
		out.Step(style.Unpause, false, "Paused {{.count}} containers in: {{.namespaces}}", out.V{"count": len(ids), "namespaces": strings.Join(namespaces, ", ")})
	}
}

func init() {
	pauseCmd.Flags().StringSliceVarP(&namespaces, "--namespaces", "n", constants.DefaultNamespaces, "namespaces to pause")
	pauseCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "If set, pause all namespaces")
	pauseCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Format to print stdout in. Options include: [text,json]")
}
