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
	"os"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	pkg_config "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	pkgutil "k8s.io/minikube/pkg/util"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Gets the status of a local kubernetes cluster",
	Long:  `Gets the status of a local kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithCode(exit.Unavailable, "Error getting client: %v", err)
		}
		defer api.Close()

		hostSt, err := cluster.GetHostStatus(api)
		if err != nil {
			exit.WithError("Error getting host status", err)
		}

		kubeletSt := state.None.String()
		kubeconfigSt := state.None.String()
		apiserverSt := state.None.String()

		if hostSt == state.Running.String() {
			clusterBootstrapper, err := GetClusterBootstrapper(api, viper.GetString(cmdcfg.Bootstrapper))
			if err != nil {
				exit.WithError("Error getting bootstrapper", err)
			}

			kubeletSt, err = clusterBootstrapper.GetKubeletStatus()
			if err != nil {
				glog.Warningf("kubelet err: %v", err)
			}

			ip, err := cluster.GetHostDriverIP(api, config.GetMachineName())
			if err != nil {
				glog.Errorln("Error host driver ip status:", err)
			}

			apiserverPort, err := pkgutil.GetPortFromKubeConfig(util.GetKubeConfigPath(), config.GetMachineName())
			if err != nil {
				// Fallback to presuming default apiserver port
				apiserverPort = pkgutil.APIServerPort
			}

			apiserverSt, err = clusterBootstrapper.GetAPIServerStatus(ip, apiserverPort)
			if err != nil {
				glog.Errorln("Error apiserver status:", err)
			}

			ks, err := pkgutil.GetKubeConfigStatus(ip, util.GetKubeConfigPath(), config.GetMachineName())
			if err != nil {
				glog.Errorln("Error kubeconfig status:", err)
			}

			if ks {
				kubeconfigSt = "Correctly Configured: pointing to minikube-vm at " + ip.String()
			} else {
				kubeconfigSt = "Misconfigured: pointing to stale minikube-vm." +
					"\nTo fix the kubectl context, run minikube update-context"
			}
		}

		var data [][]string
		data = append(data, []string{pkg_config.GetMachineName(), hostSt, kubeletSt, apiserverSt, kubeconfigSt})

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Profile", "Host", "Kubelet", "APIServer", "Kubectl"})
		table.SetAutoFormatHeaders(false)
		table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		table.SetCenterSeparator("|")
		table.AppendBulk(data) // Add Bulk Data
		table.Render()
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
