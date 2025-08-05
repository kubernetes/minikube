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
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/shell"
)

var podmanEnvTmpl = fmt.Sprintf(
	"{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}"+
		"{{ if .DockerTLSVerify }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ if .DockerCertPath }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ if .ExistingDockerHost }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .ExistingDockerHost }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubePodmanProfile }}{{ .Suffix }}"+
		"{{ .UsageHint }}",
	constants.DockerHostEnv,
	constants.DockerTLSVerifyEnv,
	constants.DockerCertPathEnv,
	constants.ExistingDockerHostEnv,
	constants.MinikubeActivePodmanEnv)

// PodmanShellConfig represents the shell config for Podman
type PodmanShellConfig struct {
	shell.Config
	DockerHost            string
	DockerTLSVerify       string
	DockerCertPath        string
	MinikubePodmanProfile string

	ExistingDockerHost    string
}

var podmanUnset bool

// podmanShellCfgSet generates context variables for "podman-env"
func podmanShellCfgSet(ec PodmanEnvConfig, envMap map[string]string) *PodmanShellConfig {
	profile := ec.profile
	const usgPlz = "To point your shell to minikube's podman docker-compatible service, run:"
	usgCmd := fmt.Sprintf("minikube -p %s podman-env", profile)
	s := &PodmanShellConfig{
		Config: *shell.CfgSet(ec.EnvConfig, usgPlz, usgCmd),
	}
	s.DockerHost = envMap[constants.DockerHostEnv]
	s.DockerTLSVerify = envMap[constants.DockerTLSVerifyEnv]
	s.DockerCertPath = envMap[constants.DockerCertPathEnv]

	s.ExistingDockerHost = envMap[constants.ExistingDockerHostEnv]

	s.MinikubePodmanProfile = envMap[constants.MinikubeActivePodmanEnv]

	return s
}

// isPodmanAvailable checks if podman command is available
func isPodmanAvailable(r command.Runner) bool {
	if _, err := r.RunCmd(exec.Command("which", "podman")); err != nil {
		return false
	}

	return true
}

// podmanEnvCmd represents the podman-env command
var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Configure environment to use minikube's Podman service",
	Long:  `Sets up Docker client env variables to use minikube's Podman Docker-compatible service.`,
	Run: func(_ *cobra.Command, _ []string) {
		sh := shell.EnvConfig{
			Shell: shell.ForceShell,
		}

		if podmanUnset {
			if err := podmanUnsetScript(PodmanEnvConfig{EnvConfig: sh}, os.Stdout); err != nil {
				exit.Error(reason.InternalEnvScript, "Error generating unset output", err)
			}
			return
		}

		if !out.IsTerminal(os.Stdout) {
			out.SetSilent(true)
			exit.SetShell(true)
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

		cr := co.Config.KubernetesConfig.ContainerRuntime
		if cr != constants.CRIO && cr != constants.Docker {
			exit.Message(reason.Usage, `The podman-env command is only compatible with the "crio" and "docker" runtimes, but this cluster was configured to use the "{{.runtime}}" runtime.`,
				out.V{"runtime": cr})
		}

		r := co.CP.Runner
		if ok := isPodmanAvailable(r); !ok {
			exit.Message(reason.EnvPodmanUnavailable, `The podman service within '{{.cluster}}' is not active`, out.V{"cluster": cname})
		}

		d := co.CP.Host.Driver
		hostIP := co.CP.IP.String()
		
		// Use Docker API compatibility - podman supports Docker API on port 2376
		port := constants.DockerDaemonPort
		noProxy := false
		
		// Check if we need to use SSH tunnel for remote access  
		sshHost := false
		if driver.NeedsPortForward(driverName) {
			sshHost = true
			sshPort, err := d.GetSSHPort()
			if err != nil {
				exit.Error(reason.IfSSHClient, "Error getting ssh port", err)
			}
			hostIP = "127.0.0.1"
			_ = sshPort // We'll use SSH tunnel if needed
		}

		ec := PodmanEnvConfig{
			EnvConfig: sh,
			profile:   cname,
			driver:    driverName,
			ssh:       sshHost,
			hostIP:    hostIP,
			port:      port,
			certsDir:  localpath.MakeMiniPath("certs"),
			noProxy:   noProxy,
		}

		if ec.Shell == "" {
			var err error
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
	ssh      bool
	hostIP   string
	port     int
	certsDir string
	noProxy  bool
}

// podmanSetScript writes out a shell-compatible 'podman-env' script
func podmanSetScript(ec PodmanEnvConfig, w io.Writer) error {
	envVars := podmanEnvVars(ec)
	return shell.SetScript(w, podmanEnvTmpl, podmanShellCfgSet(ec, envVars))
}

// podmanUnsetScript writes out a shell-compatible 'podman-env unset' script
func podmanUnsetScript(ec PodmanEnvConfig, w io.Writer) error {
	vars := podmanEnvNames(ec)
	return shell.UnsetScript(ec.EnvConfig, w, vars)
}


// podmanEnvVars gets the necessary Docker-compatible env variables for podman service
func podmanEnvVars(ec PodmanEnvConfig) map[string]string {
	var rt string
	if ec.ssh {
		rt = fmt.Sprintf("tcp://%s:%d", ec.hostIP, ec.port)
	} else {
		rt = fmt.Sprintf("tcp://%s:%d", ec.hostIP, ec.port)
	}
	
	env := map[string]string{
		constants.DockerHostEnv:           rt,
		constants.DockerTLSVerifyEnv:      "1",
		constants.DockerCertPathEnv:       ec.certsDir,
		constants.MinikubeActivePodmanEnv: ec.profile,
	}
	
	// Save existing Docker env if not already using minikube
	if os.Getenv(constants.MinikubeActivePodmanEnv) == "" {
		for _, envVar := range constants.DockerDaemonEnvs {
			if v := oci.InitialEnv(envVar); v != "" {
				key := constants.MinikubeExistingPrefix + envVar
				env[key] = v
			}
		}
	}
	return env
}

// podmanEnvNames gets the necessary Docker env variables to reset after using minikube's podman service
func podmanEnvNames(ec PodmanEnvConfig) []string {
	vars := []string{
		constants.DockerHostEnv,
		constants.DockerTLSVerifyEnv,
		constants.DockerCertPathEnv,
		constants.MinikubeActivePodmanEnv,
	}
	return vars
}

func init() {
	podmanEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	podmanEnvCmd.Flags().BoolVarP(&podmanUnset, "unset", "u", false, "Unset variables instead of setting them")
}
