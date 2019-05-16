/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"os"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/shell"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
)

const (
	envTmpl = `{{ .Prefix }}DOCKER_TLS_VERIFY{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}{{ .Prefix }}DOCKER_HOST{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}{{ .Prefix }}DOCKER_CERT_PATH{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}{{ .Prefix }}DOCKER_API_VERSION{{ .Delimiter }}{{ .DockerAPIVersion }}{{ .Suffix }}{{ if .NoProxyVar }}{{ .Prefix }}{{ .NoProxyVar }}{{ .Delimiter }}{{ .NoProxyValue }}{{ .Suffix }}{{end}}{{ .UsageHint }}`

	fishSetPfx   = "set -gx "
	fishSetSfx   = "\";\n"
	fishSetDelim = " \""

	fishUnsetPfx   = "set -e "
	fishUnsetSfx   = ";\n"
	fishUnsetDelim = ""

	psSetPfx   = "$Env:"
	psSetSfx   = "\"\n"
	psSetDelim = " = \""

	psUnsetPfx   = `Remove-Item Env:\\`
	psUnsetSfx   = "\n"
	psUnsetDelim = ""

	cmdSetPfx   = "SET "
	cmdSetSfx   = "\n"
	cmdSetDelim = "="

	cmdUnsetPfx   = "SET "
	cmdUnsetSfx   = "\n"
	cmdUnsetDelim = "="

	emacsSetPfx   = "(setenv \""
	emacsSetSfx   = "\")\n"
	emacsSetDelim = "\" \""

	emacsUnsetPfx   = "(setenv \""
	emacsUnsetSfx   = ")\n"
	emacsUnsetDelim = "\" nil"

	bashSetPfx   = "export "
	bashSetSfx   = "\"\n"
	bashSetDelim = "=\""

	bashUnsetPfx   = "unset "
	bashUnsetSfx   = "\n"
	bashUnsetDelim = ""

	nonePfx   = ""
	noneSfx   = "\n"
	noneDelim = "="
)

var usageHintMap = map[string]string{
	"bash": `# Run this command to configure your shell:
# eval $(minikube docker-env)
`,
	"fish": `# Run this command to configure your shell:
# eval (minikube docker-env)
`,
	"powershell": `# Run this command to configure your shell:
# & minikube docker-env | Invoke-Expression
`,
	"cmd": `REM Run this command to configure your shell:
REM @FOR /f "tokens=*" %i IN ('minikube docker-env') DO @%i
`,
	"emacs": `;; Run this command to configure your shell:
;; (with-temp-buffer (shell-command "minikube docker-env" (current-buffer)) (eval-buffer))
`,
}

// ShellConfig represents the shell config
type ShellConfig struct {
	Prefix           string
	Delimiter        string
	Suffix           string
	DockerCertPath   string
	DockerHost       string
	DockerTLSVerify  string
	DockerAPIVersion string
	UsageHint        string
	NoProxyVar       string
	NoProxyValue     string
}

var (
	noProxy              bool
	forceShell           string
	unset                bool
	defaultShellDetector ShellDetector
	defaultNoProxyGetter NoProxyGetter
)

// ShellDetector detects shell
type ShellDetector interface {
	GetShell(string) (string, error)
}

// LibmachineShellDetector detects shell, using libmachine
type LibmachineShellDetector struct{}

// NoProxyGetter gets the no_proxy variable
type NoProxyGetter interface {
	GetNoProxyVar() (string, string)
}

// EnvNoProxyGetter gets the no_proxy variable, using environment
type EnvNoProxyGetter struct{}

func generateUsageHint(userShell string) string {
	hint, ok := usageHintMap[userShell]
	if !ok {
		return usageHintMap["bash"]
	}
	return hint
}

func shellCfgSet(api libmachine.API) (*ShellConfig, error) {

	envMap, err := cluster.GetHostDockerEnv(api)
	if err != nil {
		return nil, err
	}

	userShell, err := defaultShellDetector.GetShell(forceShell)
	if err != nil {
		return nil, err
	}

	shellCfg := &ShellConfig{
		DockerCertPath:   envMap["DOCKER_CERT_PATH"],
		DockerHost:       envMap["DOCKER_HOST"],
		DockerTLSVerify:  envMap["DOCKER_TLS_VERIFY"],
		DockerAPIVersion: constants.DockerAPIVersion,
		UsageHint:        generateUsageHint(userShell),
	}

	if noProxy {
		host, err := api.Load(config.GetMachineName())
		if err != nil {
			return nil, errors.Wrap(err, "Error getting IP")
		}

		ip, err := host.Driver.GetIP()
		if err != nil {
			return nil, errors.Wrap(err, "Error getting host IP")
		}

		noProxyVar, noProxyValue := defaultNoProxyGetter.GetNoProxyVar()

		// add the docker host to the no_proxy list idempotently
		switch {
		case noProxyValue == "":
			noProxyValue = ip
		case strings.Contains(noProxyValue, ip):
		// ip already in no_proxy list, nothing to do
		default:
			noProxyValue = fmt.Sprintf("%s,%s", noProxyValue, ip)
		}

		shellCfg.NoProxyVar = noProxyVar
		shellCfg.NoProxyValue = noProxyValue
	}

	switch userShell {
	case "fish":
		shellCfg.Prefix = fishSetPfx
		shellCfg.Suffix = fishSetSfx
		shellCfg.Delimiter = fishSetDelim
	case "powershell":
		shellCfg.Prefix = psSetPfx
		shellCfg.Suffix = psSetSfx
		shellCfg.Delimiter = psSetDelim
	case "cmd":
		shellCfg.Prefix = cmdSetPfx
		shellCfg.Suffix = cmdSetSfx
		shellCfg.Delimiter = cmdSetDelim
	case "emacs":
		shellCfg.Prefix = emacsSetPfx
		shellCfg.Suffix = emacsSetSfx
		shellCfg.Delimiter = emacsSetDelim
	case "none":
		shellCfg.Prefix = nonePfx
		shellCfg.Suffix = noneSfx
		shellCfg.Delimiter = noneDelim
		shellCfg.UsageHint = ""
	default:
		shellCfg.Prefix = bashSetPfx
		shellCfg.Suffix = bashSetSfx
		shellCfg.Delimiter = bashSetDelim
	}

	return shellCfg, nil
}

func shellCfgUnset() (*ShellConfig, error) {

	userShell, err := defaultShellDetector.GetShell(forceShell)
	if err != nil {
		return nil, err
	}

	shellCfg := &ShellConfig{
		UsageHint: generateUsageHint(userShell),
	}

	if noProxy {
		shellCfg.NoProxyVar, shellCfg.NoProxyValue = defaultNoProxyGetter.GetNoProxyVar()
	}

	switch userShell {
	case "fish":
		shellCfg.Prefix = fishUnsetPfx
		shellCfg.Suffix = fishUnsetSfx
		shellCfg.Delimiter = fishUnsetDelim
	case "powershell":
		shellCfg.Prefix = psUnsetPfx
		shellCfg.Suffix = psUnsetSfx
		shellCfg.Delimiter = psUnsetDelim
	case "cmd":
		shellCfg.Prefix = cmdUnsetPfx
		shellCfg.Suffix = cmdUnsetSfx
		shellCfg.Delimiter = cmdUnsetDelim
	case "emacs":
		shellCfg.Prefix = emacsUnsetPfx
		shellCfg.Suffix = emacsUnsetSfx
		shellCfg.Delimiter = emacsUnsetDelim
	case "none":
		shellCfg.Prefix = nonePfx
		shellCfg.Suffix = noneSfx
		shellCfg.Delimiter = noneDelim
		shellCfg.UsageHint = ""
	default:
		shellCfg.Prefix = bashUnsetPfx
		shellCfg.Suffix = bashUnsetSfx
		shellCfg.Delimiter = bashUnsetDelim
	}

	return shellCfg, nil
}

func executeTemplateStdout(shellCfg *ShellConfig) error {
	tmpl := template.Must(template.New("envConfig").Parse(envTmpl))
	return tmpl.Execute(os.Stdout, shellCfg)
}

// GetShell detects the shell
func (LibmachineShellDetector) GetShell(userShell string) (string, error) {
	if userShell != "" {
		return userShell, nil
	}
	return shell.Detect()
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

// same as drivers.RunSSHCommandFromDriver, but allows errors
func runSSHCommandFromDriver(d drivers.Driver, command string) (string, error) {
	client, err := drivers.GetSSHClientFromDriver(d)
	if err != nil {
		return "", err
	}

	log.Debugf("About to run SSH command:\n%s", command)
	output, err := client.Output(command)
	log.Debugf("SSH cmd err, output: %v: %s", err, output)
	return output, err
}

// same as host.RunSSHCommand, but allows errors
func runSSHCommand(h *host.Host, command string) (string, error) {
	return runSSHCommandFromDriver(h.Driver, command)
}

// GetDockerActive checks if Docker is active
func GetDockerActive(host *host.Host) (bool, error) {
	statusCmd := `sudo systemctl is-active docker`
	status, err := runSSHCommand(host, statusCmd)
	// systemd returns error code on inactive
	s := strings.TrimSpace(status)
	return err == nil && s == "active", nil
}

// envCmd represents the docker-env command
var dockerEnvCmd = &cobra.Command{
	Use:   "docker-env",
	Short: "Sets up docker env variables; similar to '$(docker-machine env)'",
	Long:  `Sets up docker env variables; similar to '$(docker-machine env)'.`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()
		host, err := cluster.CheckIfHostExistsAndLoad(api, config.GetMachineName())
		if err != nil {
			exit.WithError("Error getting host", err)
		}
		if host.Driver.DriverName() == "none" {
			exit.Usage(`'none' driver does not support 'minikube docker-env' command`)
		}
		hostSt, err := cluster.GetHostStatus(api)
		if err != nil {
			exit.WithError("Error getting host status", err)
		}
		if hostSt != state.Running.String() {
			exit.WithCode(exit.Unavailable, `The docker host is currently not running`)
		}
		docker, err := GetDockerActive(host)
		if err != nil {
			exit.WithError("Error getting service status", err)
		}
		if !docker {
			exit.WithCode(exit.Unavailable, `The docker service is currently not active`)
		}

		var shellCfg *ShellConfig

		if unset {
			shellCfg, err = shellCfgUnset()
			if err != nil {
				exit.WithError("Error unsetting shell variables", err)
			}
		} else {
			shellCfg, err = shellCfgSet(api)
			if err != nil {
				exit.WithError("Error setting shell variables", err)
			}
		}

		if err := executeTemplateStdout(shellCfg); err != nil {
			exit.WithError("Error executing template", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(dockerEnvCmd)
	defaultShellDetector = &LibmachineShellDetector{}
	defaultNoProxyGetter = &EnvNoProxyGetter{}
	dockerEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add machine IP to NO_PROXY environment variable")
	dockerEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	dockerEnvCmd.Flags().BoolVarP(&unset, "unset", "u", false, "Unset variables instead of setting them")
}
