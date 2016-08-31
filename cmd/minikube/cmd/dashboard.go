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
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"

	commonutil "k8s.io/minikube/pkg/util"
)

var (
	dashboardURLMode bool
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Opens/displays the kubernetes dashboard URL for your local cluster",
	Long:  `Opens/displays the kubernetes dashboard URL for your local cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		cluster.EnsureMinikubeRunningOrExit(api)
		namespace := "kube-system"
		service := "kubernetes-dashboard"

		if err := commonutil.RetryAfter(20, func() error { return CheckService(namespace, service) }, 6*time.Second); err != nil {
			fmt.Fprintln(os.Stderr, "Could not find finalized endpoint being pointed to by %s: %s", service, err)
			util.MaybeReportErrorAndExit(err)
		}

		url, err := cluster.GetServiceURL(api, namespace, service)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "Check that minikube is running.")
			util.MaybeReportErrorAndExit(err)
		}
		if dashboardURLMode {
			fmt.Fprintln(os.Stdout, url)
		} else {
			fmt.Fprintln(os.Stdout, "Opening kubernetes dashboard in default browser...")
			browser.OpenURL(url)
		}

	},
}

func init() {
	dashboardCmd.Flags().BoolVar(&dashboardURLMode, "url", false, "Display the kubernetes dashboard in the CLI instead of opening it in the default browser")
	RootCmd.AddCommand(dashboardCmd)
}
