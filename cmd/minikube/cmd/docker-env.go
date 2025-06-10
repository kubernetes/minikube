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
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/environment"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/shell"
	"k8s.io/minikube/pkg/minikube/sshagent"
)

var (
	noProxy     bool
	sshHost     bool
	sshAdd      bool
	dockerUnset bool
)

var dockerEnvCmd = &cobra.Command{
	Use:   "docker-env",
	Short: "Configures the shell to use the container engine inside minikube (Docker, Podman, or Containerd)",
	Long: `Configures the shell to use the container engine inside minikube. 
This is useful for building images directly inside the unified container runtime environment.`,
	Run: func(_ *cobra.Command, _ []string) {
		sh, err := shell.GetShell(shell.ForceShell)
		if err != nil {
			exit.Error(reason.InternalShellDetect, "Error detecting shell", err)
		}

		cname := ClusterFlagValue()
		co, err := mustload.Running(cname)
		if err != nil {
			exit.Error(reason.InternalLoad, "Error loading cluster", err)
		}
		
		driverName := co.CP.Host.DriverName
		cr := co.Config.KubernetesConfig.ContainerRuntime

		var ec environment.EnvConfigurator
		
		if driverName == driver.Podman || cr == constants.CRIO {
			ec, err = environment.NewPodmanConfigurator(co)
		} else { // Docker and Containerd runtimes
			// The constructor now handles the containerd ssh-agent logic internally
			ec, err = environment.NewDockerConfigurator(co, sshHost, noProxy)
		}
		if err != nil {
			exit.Error(reason.InternalGuestEnvironment, "Failed to get environment config", err)
		}

		if dockerUnset {
			vars, err := ec.UnsetVars()
			if err != nil {
				exit.Error(reason.InternalCommand, "Failed to get vars to unset", err)
			}
			if err := shell.UnsetScript(shell.Config{Shell: sh}, os.Stdout, vars); err != nil {
				exit.Error(reason.InternalEnvScript, "Error generating unset script", err)
			}
			return
		}

		// --- Pre-flight checks ---
		if driverName == driver.None {
			exit.Message(reason.EnvDriverConflict, `'none' driver does not support this command`)
		}
		if len(co.Config.Nodes) > 1 {
			exit.Message(reason.EnvMultiConflict, `This command is incompatible with multi-node clusters.`)
		}
		if err := checkSupport(cr, driverName); err != nil {
			exit.Message(reason.Usage, err.Error())
		}
		
		// This logic needs to happen *after* the configurator is created, which might have started the agent
		if co.Config.KubernetesConfig.ContainerRuntime == constants.Containerd {
			sshAdd = true 
		}

		// Unified call to the interface method
		if err := ec.DisplayScript(shell.Config{Shell: sh}, os.Stdout); err != nil {
			exit.Error(reason.InternalEnvScript, "Error generating environment script", err)
		}

		if sshAdd {
			// Reload config to ensure we have the latest SSH agent PID/Socket if it was just started
			reloadedCo, err := mustload.Running(cname)
			if err != nil {
				exit.Error(reason.InternalLoad, "Error reloading cluster config for ssh-add", err)
			}

			d := reloadedCo.CP.Host.Driver
			keyPath := d.GetSSHKeyPath()
			klog.Infof("Adding %v to ssh-agent", keyPath)
			path, err := exec.LookPath("ssh-add")
			if err != nil {
				exit.Error(reason.IfSSHClient, "Error with ssh-add", err)
			}
			cmd := exec.Command(path, keyPath)
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(), fmt.Sprintf("SSH_AUTH_SOCK=%s", reloadedCo.Config.SSHAuthSock), fmt.Sprintf("SSH_AGENT_PID=%d", reloadedCo.Config.SSHAgentPID))
			if err := cmd.Run(); err != nil {
				exit.Error(reason.IfSSHClient, "Error with ssh-add", err)
			}
		}
	},
}

// checkSupport verifies if the combination of runtime and driver is supported.
func checkSupport(containerRuntime, driverName string) error {
	if containerRuntime == constants.CRIO && driverName == driver.Docker {
		return fmt.Errorf("the CRI-O runtime is not supported with the docker driver")
	}
	if containerRuntime == constants.Containerd && driverName != driver.Docker {
		return fmt.Errorf("using containerd with the 'docker-env' command is only supported with the docker driver")
	}
	return nil
}

func init() {
	dockerEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add machine IP to NO_PROXY environment variable")
	dockerEnvCmd.Flags().BoolVar(&sshHost, "ssh-host", false, "Use SSH connection instead of HTTPS (port 2376)")
	dockerEnvCmd.Flags().BoolVar(&sshAdd, "ssh-add", false, "Add SSH identity key to SSH authentication agent")
	dockerEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	dockerEnvCmd.Flags().BoolVarP(&dockerUnset, "unset", "u", false, "Unset variables instead of setting them")
}

var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Configure environment to use minikube's Podman service (DEPRECATED: use 'docker-env')",
	Long:  `Sets up podman env variables; similar to '$(podman-machine env)'.(The podman-env command is planned for removal because its functionality is now compatibly handled by docker-env, Please check: https://github.com/kubernetes/minikube/issues/20828)`,
	Run:   func(cmd *cobra.Command, args []string) {
		out.WarningT("The 'podman-env' command is deprecated and will be removed in a future version. Please use 'docker-env', which now supports all runtimes.")
		dockerEnvCmd.Run(cmd, args)
	},
}
