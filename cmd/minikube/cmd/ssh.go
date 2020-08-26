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
	"os"

	"github.com/spf13/cobra"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/exitcode"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
)

var (
	nativeSSHClient bool
)

// sshCmd represents the docker-ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Log into the minikube environment (for debugging)",
	Long:  "Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'.",
	Run: func(cmd *cobra.Command, args []string) {
		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		if co.CP.Host.DriverName == driver.None {
			exit.UsageT("'none' driver does not support 'minikube ssh' command")
		}

		var err error
		var n *config.Node
		if nodeName == "" {
			n = co.CP.Node
		} else {
			n, _, err = node.Retrieve(*co.Config, nodeName)
			if err != nil {
				exit.WithCodeT(exitcode.GuestNotFound, "Node {{.nodeName}} does not exist.", out.V{"nodeName": nodeName})
			}
		}

		err = machine.CreateSSHShell(co.API, *co.Config, *n, args, nativeSSHClient)
		if err != nil {
			// This is typically due to a non-zero exit code, so no need for flourish.
			out.ErrLn("ssh: %v", err)
			// It'd be nice if we could pass up the correct error code here :(
			os.Exit(exitcode.Failure)
		}
	},
}

func init() {
	sshCmd.Flags().Bool(nativeSSH, true, "Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'.")
	sshCmd.Flags().StringVarP(&nodeName, "node", "n", "", "The node to ssh into. Defaults to the primary control plane.")
}
