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

	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
)

var (
	urlMode bool
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Opens/displays the kubernetes dashboard URL for your local cluster",
	Long:  `Opens/displays the kubernetes dashboard URL for your local cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		url, err := cluster.GetDashboardURL(api)
		if err != nil {
			glog.Errorln("Error accessing the kubernetes dashboard (is minikube running?): Error: ", err)
			os.Exit(1)
		}
		if urlMode {
			fmt.Fprintln(os.Stdout, url)
		} else {
			fmt.Fprintln(os.Stdout, "Opening kubernetes dashboard in default browser...")
			browser.OpenURL(url)
		}
	},
}

func init() {
	dashboardCmd.Flags().BoolVarP(&urlMode, "url", "", false, "Display the kubernetes dashboard in the CLI instead of opening it in the default browser")
	RootCmd.AddCommand(dashboardCmd)
}
