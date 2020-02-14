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
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

// unpauseCmd represents the docker-pause command
var unpauseCmd = &cobra.Command{
	Use:   "unpause",
	Short: "unpause Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		cname := viper.GetString(config.MachineProfile)
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()
		cc, err := config.Load(cname)

		if err != nil && !config.IsNotExist(err) {
			exit.WithError("Error loading profile config", err)
		}

		if err != nil {
			out.ErrT(out.Meh, `"{{.name}}" profile does not exist`, out.V{"name": cname})
			os.Exit(1)
		}
		glog.Infof("config: %+v", cc)
		host, err := machine.CheckIfHostExistsAndLoad(api, cname)
		if err != nil {
			exit.WithError("Error getting host", err)
		}

		r, err := machine.CommandRunner(host)
		if err != nil {
			exit.WithError("Failed to get command runner", err)
		}

		cr, err := cruntime.New(cruntime.Config{Type: cc.KubernetesConfig.ContainerRuntime, Runner: r})
		if err != nil {
			exit.WithError("Failed runtime", err)
		}

		glog.Infof("namespaces: %v keys: %v", namespaces, viper.AllSettings())
		if allNamespaces {
			namespaces = nil //all
		} else {
			if len(namespaces) == 0 {
				exit.WithCodeT(exit.BadUsage, "Use -A to specify all namespaces")
			}
		}

		ids, err := cluster.Unpause(cr, r, namespaces)
		if err != nil {
			exit.WithError("Pause", err)
		}

		if namespaces == nil {
			out.T(out.Pause, "Unpaused kubelet and {{.count}} containers", out.V{"count": len(ids)})
		} else {
			out.T(out.Pause, "Unpaused kubelet and {{.count}} containers in: {{.namespaces}}", out.V{"count": len(ids), "namespaces": strings.Join(namespaces, ", ")})
		}

	},
}

func init() {
	unpauseCmd.Flags().StringSliceVarP(&namespaces, "--namespaces", "n", cluster.DefaultNamespaces, "namespaces to unpause")
	unpauseCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "If set, unpause all namespaces")
}
