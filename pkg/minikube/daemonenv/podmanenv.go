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

package daemonenv

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine/ssh"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/shell"
)

var podmanEnvTmpl = fmt.Sprintf("{{ .Prefix }}%s{{ .Delimiter }}{{ .VarlinkBridge }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubePodmanProfile }}{{ .Suffix }}{{ .UsageHint }}", constants.PodmanVarlinkBridgeEnv, constants.MinikubeActivePodmanEnv)

// PodmanShellConfig represents the shell config for Podman
type PodmanShellConfig struct {
	shell.Config
	VarlinkBridge         string
	MinikubePodmanProfile string
}

// podmanShellCfgSet generates context variables for "podman-env"
func podmanShellCfgSet(ec PodmanEnvConfig, envMap map[string]string) *PodmanShellConfig {
	profile := ec.Profile
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
func IsPodmanAvailable(r command.Runner) bool {
	if _, err := r.RunCmd(exec.Command("which", "varlink")); err != nil {
		return false
	}

	if _, err := r.RunCmd(exec.Command("which", "podman")); err != nil {
		return false
	}

	return true
}

// PodmanEnvConfig encapsulates all external inputs into shell generation for Podman
type PodmanEnvConfig struct {
	shell.EnvConfig
	Profile string
	Driver  string
	Client  *ssh.ExternalClient
}

// podmanSetScript writes out a shell-compatible 'podman-env' script
func PodmanSetScript(ec PodmanEnvConfig, w io.Writer) error {
	envVars := podmanEnvVars(ec)
	return shell.SetScript(ec.EnvConfig, w, podmanEnvTmpl, podmanShellCfgSet(ec, envVars))
}

// podmanUnsetScript writes out a shell-compatible 'podman-env unset' script
func PodmanUnsetScript(ec PodmanEnvConfig, w io.Writer) error {
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
		constants.PodmanVarlinkBridgeEnv:  podmanBridge(ec.Client),
		constants.MinikubeActivePodmanEnv: ec.Profile,
	}
	return env
}
