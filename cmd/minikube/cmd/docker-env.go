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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/daemonenv"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/shell"
)

var (
	noProxy     bool
	dockerUnset bool
)

// dockerenvCmd represents the docker-env command
var dockerEnvCmd = &cobra.Command{
	Use:   "docker-env",
	Short: "Configure environment to use minikube's Docker daemon",
	Long:  `Sets up docker env variables; similar to '$(docker-machine env)'.`,
	Run: func(cmd *cobra.Command, args []string) {
		sh := shell.EnvConfig{
			Shell: shell.ForceShell,
		}

		if dockerUnset {
			if err := daemonenv.DockerUnsetScript(daemonenv.DockerEnvConfig{EnvConfig: sh}, os.Stdout); err != nil {
				exit.WithError("Error generating unset output", err)
			}
			return
		}

		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		driverName := co.CP.Host.DriverName

		if driverName == driver.None {
			exit.UsageT(`'none' driver does not support 'minikube docker-env' command`)
		}

		if len(co.Config.Nodes) > 1 {
			exit.WithCodeT(exit.BadUsage, `The docker-env command is incompatible with multi-node clusters. Use the 'registry' add-on: https://minikube.sigs.k8s.io/docs/handbook/registry/`)
		}

		if co.Config.KubernetesConfig.ContainerRuntime != "docker" {
			exit.WithCodeT(exit.BadUsage, `The docker-env command is only compatible with the "docker" runtime, but this cluster was configured to use the "{{.runtime}}" runtime.`,
				out.V{"runtime": co.Config.KubernetesConfig.ContainerRuntime})
		}

		if ok := daemonenv.IsDockerActive(co.CP.Runner); !ok {
			glog.Warningf("dockerd is not active will try to restart it...")
			daemonenv.MustRestartDocker(cname, co.CP.Runner)
		}
		sh, err := shell.GetShell(shell.ForceShell)
		if err != nil {
			exit.WithError("Error detecting shell", err)
		}

		daemonenv.MaybeRestartDocker(cname, co.CP.Runner)

		port := constants.DockerDaemonPort
		if driver.NeedsPortForward(driverName) {
			_, err = oci.ForwardedPort(driverName, cname, port)
			if err != nil {
				exit.WithCodeT(exit.Failure, "Error getting port binding for '{{.driver_name}} driver: {{.error}}", out.V{"driver_name": driverName, "error": err})
			}
		}

		ec := daemonenv.DockerEnvConfig{
			EnvConfig: sh,
			Profile:   cname,
			Driver:    driverName,
			HostIP:    co.CP.IP.String(),
			Port:      port,
			CertsDir:  localpath.MakeMiniPath("certs"),
			NoProxy:   noProxy,
		}

		out, err := daemonenv.TryDockerConnectivity("docker", ec)
		if err != nil { // docker might be up but been loaded with wrong certs/config
			// to fix issues like this #8185
			glog.Warningf("couldn't connect to docker inside minikube. will try to restart dockerd service... output: %s error: %v", string(out), err)
			daemonenv.MustRestartDocker(cname, co.CP.Runner)
		}

		if dockerUnset {
			if err := daemonenv.DockerUnsetScript(ec, os.Stdout); err != nil {
				exit.WithError("Error generating unset output", err)
			}
			return
		}

		if err := daemonenv.DockerSetScript(ec, os.Stdout); err != nil {
			exit.WithError("Error generating set output", err)
		}
	},
}

func init() {
	dockerEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add machine IP to NO_PROXY environment variable")
	dockerEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	dockerEnvCmd.Flags().BoolVarP(&dockerUnset, "unset", "u", false, "Unset variables instead of setting them")
}
