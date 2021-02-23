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
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/sshutil"
)

var (
	appendKnown bool
)

// sshHostCmd represents the sshHostCmd command
var sshHostCmd = &cobra.Command{
	Use:   "ssh-host",
	Short: "Retrieve the ssh host key of the specified node",
	Long:  "Retrieve the ssh host key of the specified node.",
	Run: func(cmd *cobra.Command, args []string) {
		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		if co.CP.Host.DriverName == driver.Native {
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

		if appendKnown {
			addr, port, err := machine.GetSSHHostAddrPort(co.API, *co.Config, *n)
			if err != nil {
				out.ErrLn("GetSSHHostAddrPort: %v", err)
				os.Exit(1)
			}

			host := addr
			if port != 22 {
				host = fmt.Sprintf("[%s]:%d", addr, port)
			}
			knownHosts := filepath.Join(homedir.HomeDir(), ".ssh", "known_hosts")

			fmt.Fprintf(os.Stderr, "Host added: %s (%s)\n", knownHosts, host)
			if sshutil.KnownHost(host, knownHosts) {
				return
			}

			f, err := os.OpenFile(knownHosts, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				out.ErrLn("OpenFile: %v", err)
				os.Exit(1)
			}
			defer f.Close()

			_, err = f.WriteString(keys)
			if err != nil {
				out.ErrLn("WriteString: %v", err)
				os.Exit(1)
			}

			return
		}

		fmt.Printf("%s", keys)
	},
}

func init() {
	sshHostCmd.Flags().StringVarP(&nodeName, "node", "n", "", "The node to ssh into. Defaults to the primary control plane.")
	sshHostCmd.Flags().BoolVar(&appendKnown, "append-known", false, "Add host key to SSH known_hosts file")
}
