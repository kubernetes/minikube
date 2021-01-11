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

package config

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

var addonsImagesCmd = &cobra.Command{
	Use:     "images ADDON_NAME",
	Short:   "List image names the addon w/ADDON_NAME used. For a list of available addons use: minikube addons list",
	Long:    "List image names the addon w/ADDON_NAME used. For a list of available addons use: minikube addons list",
	Example: "minikube addons images ingress",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "usage: minikube addons images ADDON_NAME")
		}

		addon := args[0]
		// allows for additional prompting of information when enabling addons
		if conf, ok := assets.Addons[addon]; ok {
			if conf.Images != nil {
				out.Infof("{{.name}} has following images:", out.V{"name": addon})

				var tData [][]string
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Image Name", "Default Image", "Default Registry"})
				table.SetAutoFormatHeaders(true)
				table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
				table.SetCenterSeparator("|")

				for imageName, defaultImage := range conf.Images {
					tData = append(tData, []string{imageName, defaultImage, conf.Registries[imageName]})
				}

				table.AppendBulk(tData)
				table.Render()
			} else {
				out.Infof("{{.name}} doesn't have images.", out.V{"name": addon})
			}
		} else {
			out.FailureT("No such addon {{.name}}", out.V{"name": addon})
		}
	},
}

func init() {
	AddonsCmd.AddCommand(addonsImagesCmd)
}
