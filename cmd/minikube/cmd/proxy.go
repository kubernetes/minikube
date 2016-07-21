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
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/proxy"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Creates a minikube proxy that maps ports from the minikubeVM to localhost.",
	Long:  `Creates a minikube proxy that maps ports from the minikubeVM to localhost.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating minikube proxy...")
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		host, err := api.Load(constants.MachineName)
		if err != nil {
			glog.Errorln("Error getting IP: ", err)
			os.Exit(1)
		}
		ip, err := host.Driver.GetIP()
		if err != nil {
			glog.Errorln("Error getting IP: ", err)
			os.Exit(1)
		}
		if err := proxy.StartProxy(ip); err != nil {
			fmt.Println("Errors occurred deleting machine: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(proxyCmd)
}
