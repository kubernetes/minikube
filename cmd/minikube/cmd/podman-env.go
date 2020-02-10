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
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/shell"
)

var podmanEnvTmpl = fmt.Sprintf("{{ .Prefix }}%s{{ .Delimiter }}{{ .VarlinkBridge }}{{ .Suffix }}{{ .UsageHint }}", constants.PodmanVarlinkBridgeEnv)

// PodmanShellConfig represents the shell config for Podman
type PodmanShellConfig struct {
	shell.Config
	VarlinkBridge string
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

	return s
}

// isPodmanAvailable checks if Podman is available
func isPodmanAvailable(host *host.Host) (bool, error) {
	// we need both "varlink bridge" and "podman varlink"
	if _, err := host.RunSSHCommand("which varlink"); err != nil {
		return false, err
	}
	if _, err := host.RunSSHCommand("which podman"); err != nil {
		return false, err
	}
	return true, nil
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
	Short: "Sets up podman env variables; similar to '$(podman-machine env)'",
	Long:  `Sets up podman env variables; similar to '$(podman-machine env)'.`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()

		profile := viper.GetString(config.MachineProfile)
		cc, err := config.Load(profile)
		if err != nil {
			exit.WithError("Error getting config", err)
		}
		host, err := cluster.CheckIfHostExistsAndLoad(api, cc.Name)
		if err != nil {
			exit.WithError("Error getting host", err)
		}
		if host.Driver.DriverName() == driver.None {
			exit.UsageT(`'none' driver does not support 'minikube podman-env' command`)
		}

		hostSt, err := cluster.GetHostStatus(api, cc.Name)
		if err != nil {
			exit.WithError("Error getting host status", err)
		}
		if hostSt != state.Running.String() {
			exit.WithCodeT(exit.Unavailable, `'{{.profile}}' is not running`, out.V{"profile": profile})
		}
		ok, err := isPodmanAvailable(host)
		if err != nil {
			exit.WithError("Error getting service status", err)
		}

		if !ok {
			exit.WithCodeT(exit.Unavailable, `The podman service within '{{.profile}}' is not active`, out.V{"profile": profile})
		}

		client, err := createExternalSSHClient(host.Driver)
		if err != nil {
			exit.WithError("Error getting ssh client", err)
		}

		sh := shell.EnvConfig{
			Shell: shell.ForceShell,
		}
		ec := PodmanEnvConfig{
			EnvConfig: sh,
			profile:   profile,
			driver:    host.DriverName,
			client:    client,
		}

		if ec.Shell == "" {
			ec.Shell, err = shell.Detect()
			if err != nil {
				exit.WithError("Error detecting shell", err)
			}
		}

		if podmanUnset {
			if err := podmanUnsetScript(ec, os.Stdout); err != nil {
				exit.WithError("Error generating unset output", err)
			}
			return
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
	envVars, err := podmanEnvVars(ec)
	if err != nil {
		return err
	}
	return shell.SetScript(ec.EnvConfig, w, podmanEnvTmpl, podmanShellCfgSet(ec, envVars))
}

// podmanUnsetScript writes out a shell-compatible 'podman-env unset' script
func podmanUnsetScript(ec PodmanEnvConfig, w io.Writer) error {
	vars := []string{
		constants.PodmanVarlinkBridgeEnv,
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
func podmanEnvVars(ec PodmanEnvConfig) (map[string]string, error) { // nolint result 1 (error) is always nil
	env := map[string]string{
		constants.PodmanVarlinkBridgeEnv: podmanBridge(ec.client),
	}
	return env, nil
}

func init() {
	podmanEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	podmanEnvCmd.Flags().BoolVarP(&podmanUnset, "unset", "u", false, "Unset variables instead of setting them")
}
