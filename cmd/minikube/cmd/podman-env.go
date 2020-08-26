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
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/exitcode"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/shell"
)

var podmanEnvTmpl = fmt.Sprintf("{{ .Prefix }}%s{{ .Delimiter }}{{ .VarlinkBridge }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubePodmanProfile }}{{ .Suffix }}{{ .UsageHint }}", constants.PodmanVarlinkBridgeEnv, constants.MinikubeActivePodmanEnv)

// PodmanShellConfig represents the shell config for Podman
type PodmanShellConfig struct {
	shell.Config
	VarlinkBridge         string
	MinikubePodmanProfile string
}

var (
	podmanUnset bool
)

// podmanShellCfgSet generates context variables for "podman-env"
func podmanShellCfgSet(ec PodmanEnvConfig, envMap map[string]string) *PodmanShellConfig {
	profile := ec.profile
	const usgPlz = "To point your shell to minikube's podman service, run:"
	var usgCmd = fmt.Sprintf("minikube -p %s podman-env", profile)
	s := &PodmanShellConfig{
		Config: *shell.CfgSet(ec.EnvConfig, usgPlz, usgCmd),
	}
	s.VarlinkBridge = envMap[constants.PodmanVarlinkBridgeEnv]
	s.MinikubePodmanProfile = envMap[constants.MinikubeActivePodmanEnv]

	return s
}

// isPodmanAvailable checks if Podman is available
func isPodmanAvailable(r command.Runner) bool {
	if _, err := r.RunCmd(exec.Command("which", "varlink")); err != nil {
		return false
	}

	if _, err := r.RunCmd(exec.Command("which", "podman")); err != nil {
		return false
	}

	return true
}

func createExternalSSHClient(d drivers.Driver) (*ssh.ExternalClient, error) {
	sshBinaryPath, err := exec.LookPath("ssh")
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	addr, err := d.GetSSHHostname()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{}
	if d.GetSSHKeyPath() != "" {
		auth.Keys = []string{d.GetSSHKeyPath()}
	}

	return ssh.NewExternalClient(sshBinaryPath, d.GetSSHUsername(), addr, port, auth)
}

// podmanEnvCmd represents the podman-env command
var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Configure environment to use minikube's Podman service",
	Long:  `Sets up podman env variables; similar to '$(podman-machine env)'.`,
	Run: func(cmd *cobra.Command, args []string) {
		sh := shell.EnvConfig{
			Shell: shell.ForceShell,
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
			exit.WithCodeT(exitcode.ProgramUsage, `The podman-env command is incompatible with multi-node clusters. Use the 'registry' add-on: https://minikube.sigs.k8s.io/docs/handbook/registry/`)
		}

		if ok := isPodmanAvailable(co.CP.Runner); !ok {
			exit.WithCodeT(exitcode.ServiceUnavailable, `The podman service within '{{.cluster}}' is not active`, out.V{"cluster": cname})
		}

		client, err := createExternalSSHClient(co.CP.Host.Driver)
		if err != nil {
			exit.WithError("Error getting ssh client", err)
		}

		ec := PodmanEnvConfig{
			EnvConfig: sh,
			profile:   cname,
			driver:    driverName,
			client:    client,
		}

		if ec.Shell == "" {
			ec.Shell, err = shell.Detect()
			if err != nil {
				exit.WithError("Error detecting shell", err)
			}
		}

		if err := podmanSetScript(ec, os.Stdout); err != nil {
			exit.WithError("Error generating set output", err)
		}
	},
}

// PodmanEnvConfig encapsulates all external inputs into shell generation for Podman
type PodmanEnvConfig struct {
	shell.EnvConfig
	profile string
	driver  string
	client  *ssh.ExternalClient
}

// podmanSetScript writes out a shell-compatible 'podman-env' script
func podmanSetScript(ec PodmanEnvConfig, w io.Writer) error {
	envVars := podmanEnvVars(ec)
	return shell.SetScript(ec.EnvConfig, w, podmanEnvTmpl, podmanShellCfgSet(ec, envVars))
}

// podmanUnsetScript writes out a shell-compatible 'podman-env unset' script
func podmanUnsetScript(ec PodmanEnvConfig, w io.Writer) error {
	vars := []string{
		constants.PodmanVarlinkBridgeEnv,
		constants.MinikubeActivePodmanEnv,
	}
	return shell.UnsetScript(ec.EnvConfig, w, vars)
}

// podmanBridge returns the command to use in a var for accessing the podman varlink bridge over ssh
func podmanBridge(client *ssh.ExternalClient) string {
	command := []string{client.BinaryPath}
	command = append(command, client.BaseArgs...)
	command = append(command, "--", "sudo", "varlink", "-A", `\'podman varlink \\\$VARLINK_ADDRESS\'`, "bridge")
	return strings.Join(command, " ")
}

// podmanEnvVars gets the necessary podman env variables to allow the use of minikube's podman service
func podmanEnvVars(ec PodmanEnvConfig) map[string]string {
	env := map[string]string{
		constants.PodmanVarlinkBridgeEnv:  podmanBridge(ec.client),
		constants.MinikubeActivePodmanEnv: ec.profile,
	}
	return env
}

func init() {
	podmanEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	podmanEnvCmd.Flags().BoolVarP(&podmanUnset, "unset", "u", false, "Unset variables instead of setting them")
}
