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
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/service"
)

var serviceListNamespace string

// serviceListCmd represents the service list command
var serviceListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Lists the URLs for the services in your local cluster",
	Long:  `Lists the URLs for the services in your local cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()
		serviceURLs, err := service.GetServiceURLs(api, serviceListNamespace, serviceURLTemplate)
		if err != nil {
			console.Fatal("Failed to get service URL: %v", err)
			console.ErrStyle("notice", "Check that minikube is running and that you have specified the correct namespace (-n flag) if required.")
			os.Exit(exit.Unavailable)
		}

		var data [][]string
		for _, serviceURL := range serviceURLs {
			if len(serviceURL.URLs) == 0 {
				data = append(data, []string{serviceURL.Namespace, serviceURL.Name, "No node port"})
			} else {
				data = append(data, []string{serviceURL.Namespace, serviceURL.Name, strings.Join(serviceURL.URLs, "\n")})
			}

		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Namespace", "Name", "URL"})
		table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		table.SetCenterSeparator("|")
		table.AppendBulk(data) // Add Bulk Data
		table.Render()
	},
}

func init() {
	serviceListCmd.Flags().StringVarP(&serviceListNamespace, "namespace", "n", core.NamespaceAll, "The services namespace")
	serviceCmd.AddCommand(serviceListCmd)
}
