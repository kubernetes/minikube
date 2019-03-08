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
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
)

// ipCmd represents the ip command
var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Retrieves the IP address of the running cluster",
	Long:  `Retrieves the IP address of the running cluster, and writes it to STDOUT.`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()

		host, err := api.Load(config.GetMachineName())
		if err != nil {
			switch err := errors.Cause(err).(type) {
			case mcnerror.ErrHostDoesNotExist:
				exit.WithCode(exit.NoInput, "%q host does not exist, unable to show an IP", config.GetMachineName())
			default:
				exit.WithError("Error getting host", err)
			}
		}
		ip, err := host.Driver.GetIP()
		if err != nil {
			exit.WithError("Error getting IP", err)
		}
		console.OutLn(ip)
	},
}

func init() {
	RootCmd.AddCommand(ipCmd)
}
