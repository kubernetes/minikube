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

package daemonenv

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/shell"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

var dockerEnvTmpl = fmt.Sprintf("{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubeDockerdProfile }}{{ .Suffix }}{{ if .NoProxyVar }}{{ .Prefix }}{{ .NoProxyVar }}{{ .Delimiter }}{{ .NoProxyValue }}{{ .Suffix }}{{end}}{{ .UsageHint }}", constants.DockerTLSVerifyEnv, constants.DockerHostEnv, constants.DockerCertPathEnv, constants.MinikubeActiveDockerdEnv)

// DockerShellConfig represents the shell config for Docker
type DockerShellConfig struct {
	shell.Config
	DockerCertPath         string
	DockerHost             string
	DockerTLSVerify        string
	MinikubeDockerdProfile string
	NoProxyVar             string
	NoProxyValue           string
}

// dockerShellCfgSet generates context variables for "docker-env"
func dockerShellCfgSet(ec DockerEnvConfig, envMap map[string]string) *DockerShellConfig {
	profile := ec.Profile
	const usgPlz = "To point your shell to minikube's docker-daemon, run:"
	var usgCmd = fmt.Sprintf("minikube -p %s docker-env", profile)
	s := &DockerShellConfig{
		Config: *shell.CfgSet(ec.EnvConfig, usgPlz, usgCmd),
	}
	s.DockerCertPath = envMap[constants.DockerCertPathEnv]
	s.DockerHost = envMap[constants.DockerHostEnv]
	s.DockerTLSVerify = envMap[constants.DockerTLSVerifyEnv]
	s.MinikubeDockerdProfile = envMap[constants.MinikubeActiveDockerdEnv]

	if ec.NoProxy {
		noProxyVar, noProxyValue := GetNoProxyVar()

		// add the docker host to the no_proxy list idempotently
		switch {
		case noProxyValue == "":
			noProxyValue = ec.HostIP
		case strings.Contains(noProxyValue, ec.HostIP):
		// ip already in no_proxy list, nothing to do
		default:
			noProxyValue = fmt.Sprintf("%s,%s", noProxyValue, ec.HostIP)
		}

		s.NoProxyVar = noProxyVar
		s.NoProxyValue = noProxyValue
	}

	return s
}

// IsDockerActive checks if Docker is active
func IsDockerActive(r command.Runner) bool {
	return sysinit.New(r).Active("docker")
}

// MustRestartDocker restarts docker or exit with err
func MustRestartDocker(name string, runner command.Runner) {
	if err := sysinit.New(runner).Restart("docker"); err != nil {
		exit.WithCodeT(exit.Unavailable, `The Docker service within '{{.name}}' is not active`, out.V{"name": name})
	}
}

// MaybeRestartDocker tries to restart docker engine if needed
func MaybeRestartDocker(name string, runner command.Runner) {
	if ok := IsDockerActive(runner); !ok {
		glog.Warningf("dockerd is not active will try to restart it...")
		MustRestartDocker(name, runner)
	}
}

// DockerEnvConfig encapsulates all external inputs into shell generation for Docker
type DockerEnvConfig struct {
	shell.EnvConfig
	Profile  string
	Driver   string
	HostIP   string
	Port     int
	CertsDir string
	NoProxy  bool
}

// DockerSetScript writes out a shell-compatible 'docker-env' script
func DockerSetScript(ec DockerEnvConfig, w io.Writer) error {
	envVars := dockerEnvVars(ec)
	return shell.SetScript(ec.EnvConfig, w, dockerEnvTmpl, dockerShellCfgSet(ec, envVars))
}

// DockerUnsetScript writes out a shell-compatible 'docker-env unset' script
func DockerUnsetScript(ec DockerEnvConfig, w io.Writer) error {
	vars := []string{
		constants.DockerTLSVerifyEnv,
		constants.DockerHostEnv,
		constants.DockerCertPathEnv,
		constants.MinikubeActiveDockerdEnv,
	}

	if ec.NoProxy {
		k, _ := GetNoProxyVar()
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
	env := map[string]string{
		constants.DockerTLSVerifyEnv:       "1",
		constants.DockerHostEnv:            dockerURL(ec.HostIP, ec.Port),
		constants.DockerCertPathEnv:        ec.CertsDir,
		constants.MinikubeActiveDockerdEnv: ec.Profile,
	}
	return env
}

// dockerEnvVarsList gets the necessary docker env variables to allow the use of minikube's docker daemon to be used in a exec.Command
func dockerEnvVarsList(ec DockerEnvConfig) []string {
	var envVarList []string
	for k, v := range dockerEnvVars(ec) {
		envVarList = append(envVarList, fmt.Sprintf("%s=%s", k, v))
	}
	return envVarList
}
