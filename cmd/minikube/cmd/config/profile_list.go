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
	"fmt"
	"os"
	"strconv"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/exit"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all minikube profiles.",
	Long:  "Lists all minikube profiles.",
	Run: func(cmd *cobra.Command, args []string) {

		var validData [][]string

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Profile", "VM Driver", "NodeIP", "Node Port", "Kubernetes Version"})
		table.SetAutoFormatHeaders(false)
		table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		table.SetCenterSeparator("|")
		validProfiles, invalidProfiles, err := config.ListProfiles()

		for _, p := range validProfiles {
			validData = append(validData, []string{p.Name, p.Config.MachineConfig.VMDriver, p.Config.KubernetesConfig.NodeIP, strconv.Itoa(p.Config.KubernetesConfig.NodePort), p.Config.KubernetesConfig.KubernetesVersion})
		}

		table.AppendBulk(validData)
		table.Render()

		if invalidProfiles != nil {
			console.OutT(console.WarningType, "Found {{.number}} invalid profile(s) ! ", console.Arg{"number": len(invalidProfiles)})
			for _, p := range invalidProfiles {
				console.OutT(console.Empty, "\t "+p.Name)
			}
			console.OutT(console.Tip, "You can delete them using the following command(s): ")
			for _, p := range invalidProfiles {
				console.Out(fmt.Sprintf("\t $ minikube delete -p %s \n", p.Name))
			}

		}
		if err != nil {
			exit.WithCode(exit.Config, fmt.Sprintf("error loading profiles: %v", err))
		}
	},
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
}
