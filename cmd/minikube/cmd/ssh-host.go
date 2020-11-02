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

	"github.com/spf13/cobra"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

// sshHostCmd represents the sshHostCmd command
var sshHostCmd = &cobra.Command{
	Use:   "ssh-host",
	Short: "Retrieve the ssh host key of the specified node",
	Long:  "Retrieve the ssh host key of the specified node.",
	Run: func(cmd *cobra.Command, args []string) {
		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		if co.CP.Host.DriverName == driver.None {
			exit.Message(reason.Usage, "'none' driver does not support 'minikube ssh-host' command")
		}

		var err error
		var n *config.Node
		if nodeName == "" {
			n = co.CP.Node
		} else {
			n, _, err = node.Retrieve(*co.Config, nodeName)
			if err != nil {
				exit.Message(reason.GuestNodeRetrieve, "Node {{.nodeName}} does not exist.", out.V{"nodeName": nodeName})
			}
		}

		scanArgs := []string{"-t", "rsa"}

		keys, err := machine.RunSSHHostCommand(co.API, *co.Config, *n, "ssh-keyscan", scanArgs)
		if err != nil {
			// This is typically due to a non-zero exit code, so no need for flourish.
			out.ErrLn("ssh-keyscan: %v", err)
			// It'd be nice if we could pass up the correct error code here :(
			os.Exit(1)
		}

		fmt.Printf("%s", keys)
	},
}

func init() {
	sshHostCmd.Flags().StringVarP(&nodeName, "node", "n", "", "The node to ssh into. Defaults to the primary control plane.")
}
