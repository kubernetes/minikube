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
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	"k8s.io/minikube/pkg/drivers/kic/oci"
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
	Run: func(cmd *cobra.Command, args []string) {
		co := mustload.Healthy(ClusterFlagValue())
		output := strings.ToLower(profileOutput)

		serviceURLs, err := service.GetServiceURLs(co.API, co.Config.Name, serviceListNamespace, serviceURLTemplate)
		if err != nil {
			out.FatalT("Failed to get service URL: {{.error}}", out.V{"error": err})
			out.ErrT(style.Notice, "Check that minikube is running and that you have specified the correct namespace (-n flag) if required.")
			os.Exit(reason.ExSvcUnavailable)
		}

		switch output {
		case "table":
			printServicesTable(serviceURLs, co)
		case "json":
			printServicesJSON(serviceURLs, co)
		default:
			exit.Message(reason.Usage, fmt.Sprintf("invalid output format: %s. Valid values: 'table', 'json'", output))
		}
	},
}

func printServicesTable(serviceURLs service.URLs, co mustload.ClusterController) {
	var data [][]string
	for _, serviceURL := range serviceURLs {
		if len(serviceURL.URLs) == 0 {
			data = append(data, []string{serviceURL.Namespace, serviceURL.Name, "No node port"})
		} else {
			servicePortNames := strings.Join(serviceURL.PortNames, "\n")
			serviceURLs := strings.Join(serviceURL.URLs, "\n")

			// if we are running Docker on OSX we empty the internal service URLs
			if runtime.GOOS == "darwin" && co.Config.Driver == oci.Docker {
				serviceURLs = ""
			}

			data = append(data, []string{serviceURL.Namespace, serviceURL.Name, servicePortNames, serviceURLs})
		}
	}

	service.PrintServiceList(os.Stdout, data)
}

func printServicesJSON(serviceURLs service.URLs, co mustload.ClusterController) {
	processedServiceURLs := serviceURLs

	if runtime.GOOS == "darwin" && co.Config.Driver == oci.Docker {
		// To ensure we don't modify the original serviceURLs
		processedServiceURLs = make(service.URLs, len(serviceURLs))
		copy(processedServiceURLs, serviceURLs)

		for idx := range processedServiceURLs {
			processedServiceURLs[idx].URLs = make([]string, 0)
		}
	}

	jsonString, _ := json.Marshal(processedServiceURLs)
	os.Stdout.Write(jsonString)
}

func init() {
	serviceListCmd.Flags().StringVarP(&profileOutput, "output", "o", "table", "The output format. One of 'json', 'table'")
	serviceListCmd.Flags().StringVarP(&serviceListNamespace, "namespace", "n", core.NamespaceAll, "The services namespace")
	serviceCmd.AddCommand(serviceListCmd)
}
