/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

// Part of this code is heavily inspired/copied by the following file:
// github.com/docker/machine/commands/env.go

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/daemonenv"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/shell"
)

var (
	podmanUnset bool
)

// podmanEnvCmd represents the podman-env command
var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Configure environment to use minikube's Podman service",
	Long:  `Sets up podman env variables; similar to '$(podman-machine env)'.`,
	Run: func(cmd *cobra.Command, args []string) {
		sh, err := shell.GetShell(shell.ForceShell)
		if err != nil {
			exit.WithError("Error detecting shell", err)
		}
		if podmanUnset {
			if err := podmanUnsetScript(PodmanEnvConfig{EnvConfig: sh}, os.Stdout); err != nil {
				exit.WithError("Error generating unset output", err)
			}
			return
		}

		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		driverName := co.CP.Host.DriverName

		if driverName == driver.None {
			exit.UsageT(`'none' driver does not support 'minikube podman-env' command`)
		}

		if len(co.Config.Nodes) > 1 {
			exit.WithCodeT(exit.BadUsage, `The podman-env command is incompatible with multi-node clusters. Use the 'registry' add-on: https://minikube.sigs.k8s.io/docs/handbook/registry/`)
		}

		if ok := daemonenv.IsPodmanAvailable(co.CP.Runner); !ok {
			exit.WithCodeT(exit.Unavailable, `The podman service within '{{.cluster}}' is not active`, out.V{"cluster": cname})
		}

		client, err := daemonenv.CreateExternalSSHClient(co.CP.Host.Driver)
		if err != nil {
			exit.WithError("Error getting ssh client", err)
		}

		ec := daemonenv.PodmanEnvConfig{
			EnvConfig: sh,
			Profile:   cname,
			Driver:    driverName,
			Client:    client,
		}

		if ec.Shell == "" {
			ec.Shell, err = shell.Detect()
			if err != nil {
				exit.WithError("Error detecting shell", err)
			}
		}

		if err := daemonenv.PodmanSetScript(ec, os.Stdout); err != nil {
			exit.WithError("Error generating set output", err)
		}
		if podmanUnset {
			if err := daemonenv.PodmanUnsetScript(ec, os.Stdout); err != nil {
				exit.WithError("Error generating unset output", err)
			}
			return
		}
	},
}

func init() {
	podmanEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	podmanEnvCmd.Flags().BoolVarP(&podmanUnset, "unset", "u", false, "Unset variables instead of setting them")
}
