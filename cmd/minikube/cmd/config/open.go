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
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

var (
	namespace         string
	https             bool
	addonsURLMode     bool
	addonsURLFormat   string
	addonsURLTemplate *template.Template
)

const defaultAddonsFormatTemplate = "http://{{.IP}}:{{.Port}}"

var addonsOpenCmd = &cobra.Command{
	Use:   "open ADDON_NAME",
	Short: "Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list ",
	Long:  "Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list ",
	PreRun: func(cmd *cobra.Command, args []string) {
		t, err := template.New("addonsURL").Parse(addonsURLFormat)
		if err != nil {
			fmt.Fprintln(os.Stderr, "The value passed to --format is invalid:\n\n", err)
			os.Exit(1)
		}
		addonsURLTemplate = t
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: minikube addons open ADDON_NAME")
			os.Exit(1)
		}
		addonName := args[0]
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		cluster.EnsureMinikubeRunningOrExit(api, 1)
		addon, ok := assets.Addons[addonName] // validate addon input
		if !ok {
			fmt.Fprintln(os.Stderr, fmt.Sprintf(`addon '%s' is not a valid addon packaged with minikube.
To see the list of available addons run:
minikube addons list`, addonName))
			os.Exit(1)
		}
		ok, err := addon.IsEnabled()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if !ok {
			fmt.Fprintln(os.Stderr, fmt.Sprintf(`addon '%s' is currently not enabled.
To enable this addon run:
minikube addons enable %s`, addonName, addonName))
			os.Exit(1)
		}

		namespace := "kube-system"
		key := "kubernetes.io/minikube-addons-endpoint"
		if err := util.RetryAfter(20,
			func() error {
				_, err := cluster.GetServiceListByLabel(namespace, key, addonName)
				if err != nil {
					return err
				}
				return nil
			}, 6*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "Could not find endpoint for addon %s: %s\n", addonName, err)
			os.Exit(1)
		}
		serviceList, err := cluster.GetServiceListByLabel(namespace, key, addonName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting service with namespace :%s and labels %s:%s: %s", namespace, key, addonName, err)
			os.Exit(1)
		}
		for i := range serviceList.Items {
			service := serviceList.Items[i].ObjectMeta.Name
			cluster.WaitAndMaybeOpenService(api, namespace, service, addonsURLTemplate, addonsURLMode, https)

		}
	},
}

func init() {
	AddonsCmd.AddCommand(addonsOpenCmd)
	addonsOpenCmd.Flags().BoolVar(&addonsURLMode, "url", false, "Display the kubernetes addons URL in the CLI instead of opening it in the default browser")
	addonsOpenCmd.Flags().BoolVar(&https, "https", false, "Open the addons URL with https instead of http")

	addonsOpenCmd.PersistentFlags().StringVar(&addonsURLFormat, "format", defaultAddonsFormatTemplate, "Format to output addons URL in.  This format will be applied to each url individually and they will be printed one at a time.")
	AddonsCmd.AddCommand(addonsOpenCmd)
}
