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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/style"
)

var serviceListNamespace string
var profileOutput string

// serviceListCmd represents the service list command
var serviceListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Lists the URLs for the services in your local cluster",
	Long:  `Lists the URLs for the services in your local cluster`,
	Run: func(_ *cobra.Command, _ []string) {
		co := mustload.Healthy(ClusterFlagValue())
		output := strings.ToLower(profileOutput)

		serviceURLs, err := service.GetServiceURLs(co.API, co.Config.Name, serviceListNamespace, serviceURLTemplate)
		if err != nil {
			out.ErrT(style.Fatal, "Failed to get service URL - check that minikube is running and that you have specified the correct namespace (-n flag) if required: {{.error}}", out.V{"error": err})
			os.Exit(reason.ExSvcUnavailable)
		}
		serviceURLs = updatePortsAndURLs(serviceURLs, co)

		switch output {
		case "table":
			printServicesTable(serviceURLs)
		case "json":
			printServicesJSON(serviceURLs)
		default:
			exit.Message(reason.Usage, fmt.Sprintf("invalid output format: %s. Valid values: 'table', 'json'", output))
		}
	},
}

// updatePortsAndURLs sets the port name to "No node port" if a service has no URLs and removes the URLs
// if the driver needs port forwarding as the user won't be able to hit the listed URLs which could confuse them
func updatePortsAndURLs(serviceURLs service.URLs, co mustload.ClusterController) service.URLs {
	needsPortForward := driver.NeedsPortForward(co.Config.Driver)
	for i := range serviceURLs {
		if len(serviceURLs[i].URLs) == 0 {
			serviceURLs[i].PortNames = []string{"No node port"}
		} else if needsPortForward {
			serviceURLs[i].URLs = []string{}
		}
	}
	return serviceURLs
}

func printServicesTable(serviceURLs service.URLs) {
	var data [][]string
	for _, serviceURL := range serviceURLs {
		portNames := strings.Join(serviceURL.PortNames, "\n")
		urls := strings.Join(serviceURL.URLs, "\n")
		data = append(data, []string{serviceURL.Namespace, serviceURL.Name, portNames, urls})
	}

	service.PrintServiceList(os.Stdout, data)
}

func printServicesJSON(serviceURLs service.URLs) {
	jsonString, _ := json.Marshal(serviceURLs)
	os.Stdout.Write(jsonString)
}

func init() {
	serviceListCmd.Flags().StringVarP(&profileOutput, "output", "o", "table", "The output format. One of 'json', 'table'")
	serviceListCmd.Flags().StringVarP(&serviceListNamespace, "namespace", "n", core.NamespaceAll, "The services namespace")
	serviceCmd.AddCommand(serviceListCmd)
}
