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
	"strconv"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/shell"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

var envTmpl = fmt.Sprintf("{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubeDockerdProfile }}{{ .Suffix }}{{ if .NoProxyVar }}{{ .Prefix }}{{ .NoProxyVar }}{{ .Delimiter }}{{ .NoProxyValue }}{{ .Suffix }}{{end}}{{ .UsageHint }}", constants.DockerTLSVerifyEnv, constants.DockerHostEnv, constants.DockerCertPathEnv, constants.MinikubeActiveDockerdEnv)

const (
	fishSetPfx   = "set -gx "
	fishSetSfx   = "\"\n"
	fishSetDelim = " \""

	fishUnsetPfx = "set -e "
	fishUnsetSfx = "\n"

	psSetPfx   = "$Env:"
	psSetSfx   = "\"\n"
	psSetDelim = " = \""

	psUnsetPfx = `Remove-Item Env:\\`
	psUnsetSfx = "\n"

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

	bashUnsetPfx = "unset "
	bashUnsetSfx = "\n"

	nonePfx   = ""
	noneSfx   = "\n"
	noneDelim = "="
)

// ShellConfig represents the shell config
type ShellConfig struct {
	Prefix                 string
	Delimiter              string
	Suffix                 string
	DockerCertPath         string
	DockerHost             string
	DockerTLSVerify        string
	MinikubeDockerdProfile string
	UsageHint              string
	NoProxyVar             string
	NoProxyValue           string
}

var (
	noProxy              bool
	forceShell           string
	unset                bool
	defaultNoProxyGetter NoProxyGetter
)

// NoProxyGetter gets the no_proxy variable
type NoProxyGetter interface {
	GetNoProxyVar() (string, string)
}

// EnvNoProxyGetter gets the no_proxy variable, using environment
type EnvNoProxyGetter struct{}

func generateUsageHint(profile, sh string) string {
	const usgPlz = "To point your shell to minikube's docker-daemon, run:"
	var usgCmd = fmt.Sprintf("minikube -p %s docker-env", profile)
	var usageHintMap = map[string]string{
		"bash": fmt.Sprintf(`
# %s
# eval $(%s)
`, usgPlz, usgCmd),
		"fish": fmt.Sprintf(`
# %s
# eval (%s)
`, usgPlz, usgCmd),
		"powershell": fmt.Sprintf(`# %s
# & %s | Invoke-Expression
`, usgPlz, usgCmd),
		"cmd": fmt.Sprintf(`REM %s
REM @FOR /f "tokens=*" %%i IN ('%s') DO @%%i
`, usgPlz, usgCmd),
		"emacs": fmt.Sprintf(`;; %s
;; (with-temp-buffer (shell-command "%s" (current-buffer)) (eval-buffer))
`, usgPlz, usgCmd),
	}

	hint, ok := usageHintMap[sh]
	if !ok {
		return usageHintMap["bash"]
	}
	return hint
}

// shellCfgSet generates context variables for "docker-env"
func shellCfgSet(ec EnvConfig, envMap map[string]string) *ShellConfig {
	s := &ShellConfig{
		DockerCertPath:         envMap[constants.DockerCertPathEnv],
		DockerHost:             envMap[constants.DockerHostEnv],
		DockerTLSVerify:        envMap[constants.DockerTLSVerifyEnv],
		MinikubeDockerdProfile: envMap[constants.MinikubeActiveDockerdEnv],
		UsageHint:              generateUsageHint(ec.profile, ec.shell),
	}

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

	switch ec.shell {
	case "fish":
		s.Prefix = fishSetPfx
		s.Suffix = fishSetSfx
		s.Delimiter = fishSetDelim
	case "powershell":
		s.Prefix = psSetPfx
		s.Suffix = psSetSfx
		s.Delimiter = psSetDelim
	case "cmd":
		s.Prefix = cmdSetPfx
		s.Suffix = cmdSetSfx
		s.Delimiter = cmdSetDelim
	case "emacs":
		s.Prefix = emacsSetPfx
		s.Suffix = emacsSetSfx
		s.Delimiter = emacsSetDelim
	case "none":
		s.Prefix = nonePfx
		s.Suffix = noneSfx
		s.Delimiter = noneDelim
		s.UsageHint = ""
	default:
		s.Prefix = bashSetPfx
		s.Suffix = bashSetSfx
		s.Delimiter = bashSetDelim
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

// isDockerActive checks if Docker is active
func isDockerActive(d drivers.Driver) (bool, error) {
	client, err := drivers.GetSSHClientFromDriver(d)
	if err != nil {
		return false, err
	}
	output, err := client.Output("sudo systemctl is-active docker")
	if err != nil {
		return false, err
	}
	// systemd returns error code on inactive
	s := strings.TrimSpace(output)
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
			exit.UsageT(`'none' driver does not support 'minikube docker-env' command`)
		}

		hostSt, err := cluster.GetHostStatus(api, cc.Name)
		if err != nil {
			exit.WithError("Error getting host status", err)
		}
		if hostSt != state.Running.String() {
			exit.WithCodeT(exit.Unavailable, `'{{.profile}}' is not running`, out.V{"profile": profile})
		}
		ok, err := isDockerActive(host.Driver)
		if err != nil {
			exit.WithError("Error getting service status", err)
		}

		if !ok {
			exit.WithCodeT(exit.Unavailable, `The docker service within '{{.profile}}' is not active`, out.V{"profile": profile})
		}

		hostIP, err := host.Driver.GetIP()
		if err != nil {
			exit.WithError("Error getting host IP", err)
		}

		ec := EnvConfig{
			profile:  profile,
			driver:   host.DriverName,
			shell:    forceShell,
			hostIP:   hostIP,
			certsDir: localpath.MakeMiniPath("certs"),
			noProxy:  noProxy,
		}

		if ec.shell == "" {
			ec.shell, err = shell.Detect()
			if err != nil {
				exit.WithError("Error detecting shell", err)
			}
		}

		if unset {
			if err := unsetScript(ec, os.Stdout); err != nil {
				exit.WithError("Error generating unset output", err)
			}
			return
		}

		if err := setScript(ec, os.Stdout); err != nil {
			exit.WithError("Error generating set output", err)
		}
	},
}

// EnvConfig encapsulates all external inputs into shell generation
type EnvConfig struct {
	profile  string
	shell    string
	driver   string
	hostIP   string
	certsDir string
	noProxy  bool
}

// setScript writes out a shell-compatible 'docker-env' script
func setScript(ec EnvConfig, w io.Writer) error {
	tmpl := template.Must(template.New("envConfig").Parse(envTmpl))
	envVars, err := dockerEnvVars(ec)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, shellCfgSet(ec, envVars))
}

// setScript writes out a shell-compatible 'docker-env unset' script
func unsetScript(ec EnvConfig, w io.Writer) error {
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

	var sb strings.Builder
	switch ec.shell {
	case "fish":
		for _, v := range vars {
			sb.WriteString(fmt.Sprintf("%s%s%s", fishUnsetPfx, v, fishUnsetSfx))
		}
	case "powershell":
		sb.WriteString(fmt.Sprintf("%s%s%s", psUnsetPfx, strings.Join(vars, " Env:\\\\"), psUnsetSfx))
	case "cmd":
		for _, v := range vars {
			sb.WriteString(fmt.Sprintf("%s%s%s%s", cmdUnsetPfx, v, cmdUnsetDelim, cmdUnsetSfx))
		}
	case "emacs":
		for _, v := range vars {
			sb.WriteString(fmt.Sprintf("%s%s%s%s", emacsUnsetPfx, v, emacsUnsetDelim, emacsUnsetSfx))
		}
	case "none":
		sb.WriteString(fmt.Sprintf("%s%s%s", nonePfx, strings.Join(vars, " "), noneSfx))
	default:
		sb.WriteString(fmt.Sprintf("%s%s%s", bashUnsetPfx, strings.Join(vars, " "), bashUnsetSfx))
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}

// dockerURL returns a the docker endpoint URL for an ip/port pair.
func dockerURL(ip string, port int) string {
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(port)))
}

// dockerEnvVars gets the necessary docker env variables to allow the use of minikube's docker daemon
func dockerEnvVars(ec EnvConfig) (map[string]string, error) {
	env := map[string]string{
		constants.DockerTLSVerifyEnv:       "1",
		constants.DockerHostEnv:            dockerURL(ec.hostIP, constants.DockerDaemonPort),
		constants.DockerCertPathEnv:        ec.certsDir,
		constants.MinikubeActiveDockerdEnv: ec.profile,
	}

	if driver.IsKIC(ec.driver) { // for kic we need to find out what port docker allocated during creation
		port, err := oci.HostPortBinding(ec.driver, ec.profile, constants.DockerDaemonPort)
		if err != nil {
			return nil, errors.Wrapf(err, "get hostbind port for %d", constants.DockerDaemonPort)
		}
		env[constants.DockerCertPathEnv] = dockerURL(kic.DefaultBindIPV4, port)
	}
	return env, nil
}

func init() {
	defaultNoProxyGetter = &EnvNoProxyGetter{}
	dockerEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add machine IP to NO_PROXY environment variable")
	dockerEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	dockerEnvCmd.Flags().BoolVarP(&unset, "unset", "u", false, "Unset variables instead of setting them")
}
