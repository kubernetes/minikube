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
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/machine"
)

// sshCmd represents the docker-ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'",
	Long:  "Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'.",
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			console.Fatal("Error getting client: %v", err)
			os.Exit(1)
		}
		defer api.Close()
		host, err := cluster.CheckIfHostExistsAndLoad(api, config.GetMachineName())
		if err != nil {
			console.Fatal("Error getting host: %v", err)
			os.Exit(1)
		}
		if host.Driver.DriverName() == "none" {
			console.Fatal(`'none' driver does not support 'minikube ssh' command`)
			os.Exit(0)
		}
		err = cluster.CreateSSHShell(api, args)
		if err != nil {
			console.Fatal("Error creating SSH shell: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(sshCmd)
}
