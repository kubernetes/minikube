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
	"text/template"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/service"
)

var (
	https             bool
	addonsURLMode     bool
	addonsURLFormat   string
	addonsURLTemplate *template.Template
	wait              int
	interval          int
)

const defaultAddonsFormatTemplate = "http://{{.IP}}:{{.Port}}"

var addonsOpenCmd = &cobra.Command{
	Use:   "open ADDON_NAME",
	Short: "Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list ",
	Long:  "Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list ",
	PreRun: func(cmd *cobra.Command, args []string) {
		t, err := template.New("addonsURL").Parse(addonsURLFormat)
		if err != nil {
			exit.UsageT("The value passed to --format is invalid: {{.error}}", out.V{"error": err})
		}
		addonsURLTemplate = t
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.UsageT("usage: minikube addons open ADDON_NAME")
		}
		addonName := args[0]
		// TODO(r2d4): config should not reference API, pull this out
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()

		if !cluster.IsMinikubeRunning(api) {
			os.Exit(1)
		}
		addon, ok := assets.Addons[addonName] // validate addon input
		if !ok {
			exit.WithCodeT(exit.Data, `addon '{{.name}}' is not a valid addon packaged with minikube.
To see the list of available addons run:
minikube addons list`, out.V{"name": addonName})
		}
		ok, err = addon.IsEnabled()
		if err != nil {
			exit.WithError("IsEnabled failed", err)
		}
		if !ok {
			exit.WithCodeT(exit.Unavailable, `addon '{{.name}}' is currently not enabled.
To enable this addon run:
minikube addons enable {{.name}}`, out.V{"name": addonName})
		}

		namespace := "kube-system"
		key := "kubernetes.io/minikube-addons-endpoint"

		serviceList, err := service.GetServiceListByLabel(namespace, key, addonName)
		if err != nil {
			exit.WithCodeT(exit.Unavailable, "Error getting service with namespace: {{.namespace}} and labels {{.labelName}}:{{.addonName}}: {{.error}}", out.V{"namespace": namespace, "labelName": key, "addonName": addonName, "error": err})
		}
		if len(serviceList.Items) == 0 {
			exit.WithCodeT(exit.Config, `This addon does not have an endpoint defined for the 'addons open' command.
You can add one by annotating a service with the label {{.labelName}}:{{.addonName}}`, out.V{"labelName": key, "addonName": addonName})
		}
		for i := range serviceList.Items {
			svc := serviceList.Items[i].ObjectMeta.Name
			if err := service.WaitAndMaybeOpenService(api, namespace, svc, addonsURLTemplate, addonsURLMode, https, wait, interval); err != nil {
				exit.WithCodeT(exit.Unavailable, "Wait failed: {{.error}}", out.V{"error": err})
			}
		}
	},
}

func init() {
	addonsOpenCmd.Flags().BoolVar(&addonsURLMode, "url", false, "Display the kubernetes addons URL in the CLI instead of opening it in the default browser")
	addonsOpenCmd.Flags().BoolVar(&https, "https", false, "Open the addons URL with https instead of http")
	addonsOpenCmd.Flags().IntVar(&wait, "wait", service.DefaultWait, "Amount of time to wait for service in seconds")
	addonsOpenCmd.Flags().IntVar(&interval, "interval", service.DefaultInterval, "The time interval for each check that wait performs in seconds")
	addonsOpenCmd.PersistentFlags().StringVar(&addonsURLFormat, "format", defaultAddonsFormatTemplate, "Format to output addons URL in.  This format will be applied to each url individually and they will be printed one at a time.")
	AddonsCmd.AddCommand(addonsOpenCmd)
}
