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
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/service"

	commonutil "k8s.io/minikube/pkg/util"
)

var (
	dashboardURLMode      bool
	dashboardProxyMode    bool
	dashboardProxyAddress string
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Opens/displays the kubernetes dashboard URL for your local cluster",
	Long:  `Opens/displays the kubernetes dashboard URL for your local cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		cluster.EnsureMinikubeRunningOrExit(api, 1)
		namespace := "kube-system"
		svc := "kubernetes-dashboard"

		if err = commonutil.RetryAfter(20, func() error { return service.CheckService(namespace, svc) }, 6*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "Could not find finalized endpoint being pointed to by %s: %s\n", svc, err)
			os.Exit(1)
		}

		urls, err := service.GetServiceURLsForService(api, namespace, svc, template.Must(template.New("dashboardServiceFormat").Parse(defaultServiceFormatTemplate)))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "Check that minikube is running.")
			os.Exit(1)
		}
		if len(urls) == 0 {
			errMsg := "There appears to be no url associated with dashboard, this is not expected, exiting"
			glog.Infoln(errMsg)
			os.Exit(1)
		}
		url := urls[0]
		if dashboardProxyMode {
			proxyurl, err := commonutil.Proxy(dashboardProxyAddress, url)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			url = proxyurl
		}
		if dashboardURLMode {
			fmt.Fprintln(os.Stdout, url)
		} else {
			fmt.Fprintln(os.Stdout, "Opening kubernetes dashboard in default browser...")
			browser.OpenURL(url)
		}
		if dashboardProxyMode {
			commonutil.RunTillBreak()
		}
	},
}

func init() {
	dashboardCmd.Flags().BoolVar(&dashboardURLMode, "url", false, "Display the kubernetes dashboard in the CLI instead of opening it in the default browser")
	dashboardCmd.Flags().BoolVar(&dashboardProxyMode, "proxy", false, "Keeps minikube running as a proxy server and rewrites the URL so that it refers to the proxy")
	dashboardCmd.Flags().StringVar(&dashboardProxyAddress, "proxyaddress", ":0", "Listen on a specific IP address and/or port. The format is [host]:port. Port 0 picks a random port (default: :0).")
	RootCmd.AddCommand(dashboardCmd)
}
