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
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	kcfg "k8s.io/minikube/pkg/util/kubeconfig"
)

// updateContextCmd represents the update-context command
var updateContextCmd = &cobra.Command{
	Use:   "update-context",
	Short: "Verify the IP address of the running cluster in kubeconfig.",
	Long: `Retrieves the IP address of the running cluster, checks it
			with IP in kubeconfig, and corrects kubeconfig if incorrect.`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %v\n", err)
			os.Exit(1)
		}
		defer api.Close()
		machineName := config.GetMachineName()
		ip, err := cluster.GetHostDriverIP(api, machineName)
		if err != nil {
			glog.Errorln("Error host driver ip status:", err)
			cmdUtil.MaybeReportErrorAndExit(err)
		}
		ok, err := kcfg.UpdateKubeconfigIP(ip, constants.KubeconfigPath, machineName)
		if err != nil {
			glog.Errorln("Error kubeconfig status:", err)
			cmdUtil.MaybeReportErrorAndExit(err)
		}
		if ok {
			fmt.Println("Reconfigured kubeconfig IP, now pointing at " + ip.String())
		} else {
			fmt.Println("Kubeconfig IP correctly configured, pointing at " + ip.String())
		}

	},
}

func init() {
	RootCmd.AddCommand(updateContextCmd)
}
