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
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/shell"
)

var podmanEnv1Tmpl = fmt.Sprintf("{{ .Prefix }}%s{{ .Delimiter }}{{ .VarlinkBridge }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubePodmanProfile }}{{ .Suffix }}{{ .UsageHint }}", constants.PodmanVarlinkBridgeEnv, constants.MinikubeActivePodmanEnv)

var podmanEnv2Tmpl = fmt.Sprintf("{{ .Prefix }}%s{{ .Delimiter }}{{ .ContainerHost }}{{ .Suffix }}{{ if .ContainerSSHKey }}{{ .Prefix }}%s{{ .Delimiter }}{{ .ContainerSSHKey}}{{ .Suffix }}{{ end }}{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubePodmanProfile }}{{ .Suffix }}{{ .UsageHint }}", constants.PodmanContainerHostEnv, constants.PodmanContainerSSHKeyEnv, constants.MinikubeActivePodmanEnv)

// PodmanShellConfig represents the shell config for Podman
type PodmanShellConfig struct {
	shell.Config
	VarlinkBridge         string
	ContainerHost         string
	ContainerSSHKey       string
	MinikubePodmanProfile string
}

var podmanUnset bool

// podmanShellCfgSet generates context variables for "podman-env"
func podmanShellCfgSet(ec PodmanEnvConfig, envMap map[string]string) *PodmanShellConfig {
	profile := ec.profile
	const usgPlz = "To point your shell to minikube's podman service, run:"
	usgCmd := fmt.Sprintf("minikube -p %s podman-env", profile)
	s := &PodmanShellConfig{
		Config: *shell.CfgSet(ec.EnvConfig, usgPlz, usgCmd),
	}
	s.VarlinkBridge = envMap[constants.PodmanVarlinkBridgeEnv]
	s.ContainerHost = envMap[constants.PodmanContainerHostEnv]
	s.ContainerSSHKey = envMap[constants.PodmanContainerSSHKeyEnv]
	s.MinikubePodmanProfile = envMap[constants.MinikubeActivePodmanEnv]

	return s
}

// isVarlinkAvailable checks if varlink command is available
func isVarlinkAvailable(r command.Runner) bool {
	if _, err := r.RunCmd(exec.Command("which", "varlink")); err != nil {
		return false
	}

	return true
}

// isPodmanAvailable checks if podman command is available
func isPodmanAvailable(r command.Runner) bool {
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
				exit.Error(reason.InternalEnvScript, "Error generating unset output", err)
			}
			return
		}

		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		driverName := co.CP.Host.DriverName

		if driverName == driver.None {
			exit.Message(reason.Usage, `'none' driver does not support 'minikube podman-env' command`)
		}

		if len(co.Config.Nodes) > 1 {
			exit.Message(reason.Usage, `The podman-env command is incompatible with multi-node clusters. Use the 'registry' add-on: https://minikube.sigs.k8s.io/docs/handbook/registry/`)
		}

		r := co.CP.Runner
		if ok := isPodmanAvailable(r); !ok {
			exit.Message(reason.EnvPodmanUnavailable, `The podman service within '{{.cluster}}' is not active`, out.V{"cluster": cname})
		}

		varlink := isVarlinkAvailable(r)

		d := co.CP.Host.Driver
		client, err := createExternalSSHClient(d)
		if err != nil {
			exit.Error(reason.IfSSHClient, "Error getting ssh client", err)
		}

		hostname, err := d.GetSSHHostname()
		if err != nil {
			exit.Error(reason.IfSSHClient, "Error getting ssh client", err)
		}

		port, err := d.GetSSHPort()
		if err != nil {
			exit.Error(reason.IfSSHClient, "Error getting ssh client", err)
		}

		ec := PodmanEnvConfig{
			EnvConfig: sh,
			profile:   cname,
			driver:    driverName,
			varlink:   varlink,
			client:    client,
			username:  d.GetSSHUsername(),
			hostname:  hostname,
			port:      port,
			keypath:   d.GetSSHKeyPath(),
		}

		if ec.Shell == "" {
			ec.Shell, err = shell.Detect()
			if err != nil {
				exit.Error(reason.InternalShellDetect, "Error detecting shell", err)
			}
		}

		if err := podmanSetScript(ec, os.Stdout); err != nil {
			exit.Error(reason.InternalEnvScript, "Error generating set output", err)
		}
	},
}

// PodmanEnvConfig encapsulates all external inputs into shell generation for Podman
type PodmanEnvConfig struct {
	shell.EnvConfig
	profile  string
	driver   string
	varlink  bool
	client   *ssh.ExternalClient
	username string
	hostname string
	port     int
	keypath  string
}

// podmanSetScript writes out a shell-compatible 'podman-env' script
func podmanSetScript(ec PodmanEnvConfig, w io.Writer) error {
	var podmanEnvTmpl string
	if ec.varlink {
		podmanEnvTmpl = podmanEnv1Tmpl
	} else {
		podmanEnvTmpl = podmanEnv2Tmpl
	}
	envVars := podmanEnvVars(ec)
	return shell.SetScript(ec.EnvConfig, w, podmanEnvTmpl, podmanShellCfgSet(ec, envVars))
}

// podmanUnsetScript writes out a shell-compatible 'podman-env unset' script
func podmanUnsetScript(ec PodmanEnvConfig, w io.Writer) error {
	vars := podmanEnvNames(ec)
	return shell.UnsetScript(ec.EnvConfig, w, vars)
}

// podmanBridge returns the command to use in a var for accessing the podman varlink bridge over ssh
func podmanBridge(client *ssh.ExternalClient) string {
	command := []string{client.BinaryPath}
	command = append(command, client.BaseArgs...)
	command = append(command, "--", "sudo", "varlink", "-A", `\'podman varlink \\\$VARLINK_ADDRESS\'`, "bridge")
	return strings.Join(command, " ")
}

// podmanURL returns the url to use in a var for accessing the podman socket over ssh
func podmanURL(username string, hostname string, port int) string {
	path := "/run/podman/podman.sock"
	return fmt.Sprintf("ssh://%s@%s:%d%s", username, hostname, port, path)
}

// podmanEnvVars gets the necessary podman env variables to allow the use of minikube's podman service
func podmanEnvVars(ec PodmanEnvConfig) map[string]string {
	// podman v1
	env1 := map[string]string{
		constants.PodmanVarlinkBridgeEnv: podmanBridge(ec.client),
	}
	// podman v2
	env2 := map[string]string{
		constants.PodmanContainerHostEnv:   podmanURL(ec.username, ec.hostname, ec.port),
		constants.PodmanContainerSSHKeyEnv: ec.keypath,
	}
	//common
	env0 := map[string]string{
		constants.MinikubeActivePodmanEnv: ec.profile,
	}

	var env map[string]string
	if ec.varlink {
		env = env1
	} else {
		env = env2
	}
	for k, v := range env0 {
		env[k] = v
	}
	return env
}

// podmanEnvNames gets the necessary podman env variables to reset after using minikube's podman service
func podmanEnvNames(ec PodmanEnvConfig) []string {
	// podman v1
	vars1 := []string{
		constants.PodmanVarlinkBridgeEnv,
	}
	// podman v2
	vars2 := []string{
		constants.PodmanContainerHostEnv,
		constants.PodmanContainerSSHKeyEnv,
	}
	// common
	vars0 := []string{
		constants.MinikubeActivePodmanEnv,
	}

	var vars []string
	if ec.client != nil || ec.hostname != "" {
		// getting ec.varlink needs a running machine
		if ec.varlink {
			vars = vars1
		} else {
			vars = vars2
		}
	} else {
		// just unset *all* of the variables instead
		vars = vars1
		vars = append(vars, vars2...)
	}
	vars = append(vars, vars0...)
	return vars
}

func init() {
	podmanEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	podmanEnvCmd.Flags().BoolVarP(&podmanUnset, "unset", "u", false, "Unset variables instead of setting them")
}
