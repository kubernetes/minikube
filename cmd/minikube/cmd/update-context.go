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
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
)

// updateContextCmd represents the update-context command
var updateContextCmd = &cobra.Command{
	Use:   "update-context",
	Short: "Verify the IP address of the running cluster in kubeconfig.",
	Long: `Retrieves the IP address of the running cluster, checks it
			with IP in kubeconfig, and corrects kubeconfig if incorrect.`,
	Run: func(cmd *cobra.Command, args []string) {
		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		updated, err := kubeconfig.UpdateIP(co.DriverIP, cname, kubeconfig.PathFromEnv())
		if err != nil {
			exit.WithError("update config", err)
		}
		if updated {
			out.T(out.Celebrate, "{{.cluster}} IP has been updated to point at {{.ip}}", out.V{"cluster": cname, "ip": co.DriverIP})
		} else {
			out.T(out.Meh, "{{.cluster}} IP was already correctly configured for {{.ip}}", out.V{"cluster": cname, "ip": co.DriverIP})
		}

	},
}
