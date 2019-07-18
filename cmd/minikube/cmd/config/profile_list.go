/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package config

import (
	"os"
	"strconv"

	"k8s.io/minikube/pkg/minikube/config"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all minikube profiles",
	Long:  "Lists all valid minikube profiles",
	Run: func(cmd *cobra.Command, args []string) {

		var tData [][]string

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Profile", "VM Driver", "NodeIP", "Node Port", "Kubernetes Version"})
		table.SetAutoFormatHeaders(false)
		table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		table.SetCenterSeparator("|")

		for _, p := range config.AllProfiles() {
			tData = append(tData, []string{p.Name, p.Config.MachineConfig.VMDriver, p.Config.KubernetesConfig.NodeIP, strconv.Itoa(p.Config.KubernetesConfig.NodePort), p.Config.KubernetesConfig.KubernetesVersion})
		}

		table.AppendBulk(tData) // Add Bulk Data
		table.Render()

	},
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
}
