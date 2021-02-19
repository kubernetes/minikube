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
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

// updateContextCmd represents the update-context command
var updateContextCmd = &cobra.Command{
	Use:   "update-context",
	Short: "Update kubeconfig in case of an IP or port change",
	Long: `Retrieves the IP address of the running cluster, checks it
			with IP in kubeconfig, and corrects kubeconfig if incorrect.`,
	Run: func(cmd *cobra.Command, args []string) {
		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		//	cluster extension metada for kubeconfig

		updated, err := kubeconfig.UpdateEndpoint(cname, co.CP.Hostname, co.CP.Port, kubeconfig.PathFromEnv(), kubeconfig.NewExtension())
		if err != nil {
			exit.Error(reason.HostKubeconfigUpdate, "update config", err)
		}
		if updated {
			out.Step(style.Celebrate, `"{{.context}}" context has been updated to point to {{.hostname}}:{{.port}}`, out.V{"context": cname, "hostname": co.CP.Hostname, "port": co.CP.Port})
		} else {
			out.Step(style.Meh, `No changes required for the "{{.context}}" context`, out.V{"context": cname})
		}

		if err := kubeconfig.SetCurrentContext(cname, kubeconfig.PathFromEnv()); err != nil {
			out.ErrT(style.Sad, `Error while setting kubectl current context:  {{.error}}`, out.V{"error": err})
		} else {
			out.Step(style.Kubectl, `Current context is "{{.context}}"`, out.V{"context": cname})
		}
	},
}
