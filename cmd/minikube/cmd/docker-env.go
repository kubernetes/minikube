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
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

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
	"k8s.io/minikube/pkg/minikube/sysinit"
)

var dockerSetEnvTmpl = fmt.Sprintf(
	"{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}"+
		"{{ if .ExistingDockerTLSVerify }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .ExistingDockerTLSVerify }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ if .ExistingDockerHost }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .ExistingDockerHost }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ if .ExistingDockerCertPath }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .ExistingDockerCertPath }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubeDockerdProfile }}{{ .Suffix }}"+
		"{{ if .NoProxyVar }}"+
		"{{ .Prefix }}{{ .NoProxyVar }}{{ .Delimiter }}{{ .NoProxyValue }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ .UsageHint }}",
	constants.DockerTLSVerifyEnv,
	constants.DockerHostEnv,
	constants.DockerCertPathEnv,
	constants.ExistingDockerTLSVerifyEnv,
	constants.ExistingDockerHostEnv,
	constants.ExistingDockerCertPathEnv,
	constants.MinikubeActiveDockerdEnv)

// DockerShellConfig represents the shell config for Docker
type DockerShellConfig struct {
	shell.Config
	DockerCertPath         string
	DockerHost             string
	DockerTLSVerify        string
	MinikubeDockerdProfile string
	NoProxyVar             string
	NoProxyValue           string

	ExistingDockerCertPath  string
	ExistingDockerHost      string
	ExistingDockerTLSVerify string
}

var (
	noProxy              bool
	dockerUnset          bool
	defaultNoProxyGetter NoProxyGetter
)

// NoProxyGetter gets the no_proxy variable
type NoProxyGetter interface {
	GetNoProxyVar() (string, string)
}

// EnvNoProxyGetter gets the no_proxy variable, using environment
type EnvNoProxyGetter struct{}

// dockerShellCfgSet generates context variables for "docker-env"
func dockerShellCfgSet(ec DockerEnvConfig, envMap map[string]string) *DockerShellConfig {
	profile := ec.profile
	const usgPlz = "To point your shell to minikube's docker-daemon, run:"
	usgCmd := fmt.Sprintf("minikube -p %s docker-env", profile)
	s := &DockerShellConfig{
		Config: *shell.CfgSet(ec.EnvConfig, usgPlz, usgCmd),
	}
	s.DockerCertPath = envMap[constants.DockerCertPathEnv]
	s.DockerHost = envMap[constants.DockerHostEnv]
	s.DockerTLSVerify = envMap[constants.DockerTLSVerifyEnv]

	s.ExistingDockerCertPath = envMap[constants.ExistingDockerCertPathEnv]
	s.ExistingDockerHost = envMap[constants.ExistingDockerHostEnv]
	s.ExistingDockerTLSVerify = envMap[constants.ExistingDockerTLSVerifyEnv]

	s.MinikubeDockerdProfile = envMap[constants.MinikubeActiveDockerdEnv]

	if ec.noProxy {
		noProxyVar, noProxyValue := defaultNoProxyGetter.GetNoProxyVar()

		// add the docker host to the no_proxy list idempotently
		switch {
		case noProxyValue == "":
			noProxyValue = ec.hostIP
		case strings.Contains(noProxyValue, ec.hostIP):
		// ip already in no_proxy list, nothing to do
		default:
			noProxyValue = fmt.Sprintf("%s,%s", noProxyValue, ec.hostIP)
		}

		s.NoProxyVar = noProxyVar
		s.NoProxyValue = noProxyValue
	}

	return s
}

// GetNoProxyVar gets the no_proxy var
func (EnvNoProxyGetter) GetNoProxyVar() (string, string) {
	// first check for an existing lower case no_proxy var
	noProxyVar := "no_proxy"
	noProxyValue := os.Getenv("no_proxy")

	// otherwise default to allcaps HTTP_PROXY
	if noProxyValue == "" {
		noProxyVar = "NO_PROXY"
		noProxyValue = os.Getenv("NO_PROXY")
	}
	return noProxyVar, noProxyValue
}

// ensureDockerd ensures dockerd inside minikube is running before a docker-env  command
func ensureDockerd(name string, runner command.Runner) {
	if ok := isDockerActive(runner); ok {
		return
	}

	// Docker Docs: https://docs.docker.com/config/containers/live-restore
	//   On Linux, you can avoid a restart (and avoid any downtime for your containers) by reloading the Docker daemon.
	klog.Warningf("dockerd is not active will try to reload it...")
	if err := sysinit.New(runner).Reload("docker"); err != nil {
		klog.Warningf("will try to restar docker because reload failed: %v", err)
		if err := sysinit.New(runner).Restart("docker"); err != nil {
			exit.Message(reason.RuntimeRestart, `The Docker service within '{{.name}}' is not active`, out.V{"name": name})
		}
		// if we get to the point that we have to restart docker (instead of reload)
		// will need to wait for apisever container to come up, this usually takes 5 seconds
		// verifying apisever using kverify would add code complexity for a rare case.
		klog.Warningf("waiting 5 seconds to ensure apisever container is up...")
		time.Sleep(time.Second * 5)
	}
}

// isDockerActive checks if Docker is active
func isDockerActive(r command.Runner) bool {
	return sysinit.New(r).Active("docker")
}

// dockerEnvCmd represents the docker-env command
var dockerEnvCmd = &cobra.Command{
	Use:   "docker-env",
	Short: "Configure environment to use minikube's Docker daemon",
	Long:  `Sets up docker env variables; similar to '$(docker-machine env)'.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		shl := shell.ForceShell
		if shl == "" {
			shl, err = shell.Detect()
			if err != nil {
				exit.Error(reason.InternalShellDetect, "Error detecting shell", err)
			}
		}
		sh := shell.EnvConfig{
			Shell: shl,
		}

		if dockerUnset {
			if err := dockerUnsetScript(DockerEnvConfig{EnvConfig: sh}, os.Stdout); err != nil {
				exit.Error(reason.InternalEnvScript, "Error generating unset output", err)
			}
			return
		}

		cname := ClusterFlagValue()
		co := mustload.Running(cname)
		driverName := co.CP.Host.DriverName

		if driverName == driver.None {
			exit.Message(reason.EnvDriverConflict, `'none' driver does not support 'minikube docker-env' command`)
		}

		if len(co.Config.Nodes) > 1 {
			exit.Message(reason.EnvMultiConflict, `The docker-env command is incompatible with multi-node clusters. Use the 'registry' add-on: https://minikube.sigs.k8s.io/docs/handbook/registry/`)
		}

		if co.Config.KubernetesConfig.ContainerRuntime != "docker" {
			exit.Message(reason.Usage, `The docker-env command is only compatible with the "docker" runtime, but this cluster was configured to use the "{{.runtime}}" runtime.`,
				out.V{"runtime": co.Config.KubernetesConfig.ContainerRuntime})
		}

		ensureDockerd(cname, co.CP.Runner)

		port := constants.DockerDaemonPort
		if driver.NeedsPortForward(driverName) {
			port, err = oci.ForwardedPort(driverName, cname, port)
			if err != nil {
				exit.Message(reason.DrvPortForward, "Error getting port binding for '{{.driver_name}} driver: {{.error}}", out.V{"driver_name": driverName, "error": err})
			}
		}

		ec := DockerEnvConfig{
			EnvConfig: sh,
			profile:   cname,
			driver:    driverName,
			hostIP:    co.CP.IP.String(),
			port:      port,
			certsDir:  localpath.MakeMiniPath("certs"),
			noProxy:   noProxy,
		}

		dockerPath, err := exec.LookPath("docker")
		if err != nil {
			klog.Warningf("Unable to find docker in path - skipping connectivity check: %v", err)
			dockerPath = ""
		}

		if dockerPath != "" {
			out, err := tryDockerConnectivity("docker", ec)
			if err != nil { // docker might be up but been loaded with wrong certs/config
				// to fix issues like this #8185
				klog.Warningf("couldn't connect to docker inside minikube. will try to restart dockerd service... output: %s error: %v", string(out), err)
				esnureDockerd(cname, co.CP.Runner)
			}
		}

		if err := dockerSetScript(ec, os.Stdout); err != nil {
			exit.Error(reason.InternalDockerScript, "Error generating set output", err)
		}
	},
}

// DockerEnvConfig encapsulates all external inputs into shell generation for Docker
type DockerEnvConfig struct {
	shell.EnvConfig
	profile  string
	driver   string
	hostIP   string
	port     int
	certsDir string
	noProxy  bool
}

// dockerSetScript writes out a shell-compatible 'docker-env' script
func dockerSetScript(ec DockerEnvConfig, w io.Writer) error {
	envVars := dockerEnvVars(ec)
	return shell.SetScript(ec.EnvConfig, w, dockerSetEnvTmpl, dockerShellCfgSet(ec, envVars))
}

// dockerSetScript writes out a shell-compatible 'docker-env unset' script
func dockerUnsetScript(ec DockerEnvConfig, w io.Writer) error {
	vars := []string{
		constants.DockerTLSVerifyEnv,
		constants.DockerHostEnv,
		constants.DockerCertPathEnv,
		constants.MinikubeActiveDockerdEnv,
	}

	if ec.noProxy {
		k, _ := defaultNoProxyGetter.GetNoProxyVar()
		if k != "" {
			vars = append(vars, k)
		}
	}
	return shell.UnsetScript(ec.EnvConfig, w, vars)
}

// dockerURL returns a the docker endpoint URL for an ip/port pair.
func dockerURL(ip string, port int) string {
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(port)))
}

// dockerEnvVars gets the necessary docker env variables to allow the use of minikube's docker daemon
func dockerEnvVars(ec DockerEnvConfig) map[string]string {
	rt := map[string]string{
		constants.DockerTLSVerifyEnv:       "1",
		constants.DockerHostEnv:            dockerURL(ec.hostIP, ec.port),
		constants.DockerCertPathEnv:        ec.certsDir,
		constants.MinikubeActiveDockerdEnv: ec.profile,
	}
	if os.Getenv(constants.MinikubeActiveDockerdEnv) == "" {
		for _, env := range constants.DockerDaemonEnvs {
			if v := oci.InitialEnv(env); v != "" {
				key := constants.MinikubeExistingPrefix + env
				rt[key] = v
			}
		}
	}
	return rt
}

// dockerEnvVarsList gets the necessary docker env variables to allow the use of minikube's docker daemon to be used in a exec.Command
func dockerEnvVarsList(ec DockerEnvConfig) []string {
	return []string{
		fmt.Sprintf("%s=%s", constants.DockerTLSVerifyEnv, "1"),
		fmt.Sprintf("%s=%s", constants.DockerHostEnv, dockerURL(ec.hostIP, ec.port)),
		fmt.Sprintf("%s=%s", constants.DockerCertPathEnv, ec.certsDir),
		fmt.Sprintf("%s=%s", constants.MinikubeActiveDockerdEnv, ec.profile),
	}
}

// tryDockerConnectivity will try to connect to docker env from user's POV to detect the problem if it needs reset or not
func tryDockerConnectivity(bin string, ec DockerEnvConfig) ([]byte, error) {
	c := exec.Command(bin, "version", "--format={{.Server}}")
	c.Env = append(os.Environ(), dockerEnvVarsList(ec)...)
	klog.Infof("Testing Docker connectivity with: %v", c)
	return c.CombinedOutput()
}

func init() {
	defaultNoProxyGetter = &EnvNoProxyGetter{}
	dockerEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add machine IP to NO_PROXY environment variable")
	dockerEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	dockerEnvCmd.Flags().BoolVarP(&dockerUnset, "unset", "u", false, "Unset variables instead of setting them")
}
