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

	apiWait "k8s.io/apimachinery/pkg/util/wait"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
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

const minLogCheckTime = 60 * time.Second

var dockerEnvTCPTmpl = fmt.Sprintf(
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
var dockerEnvSSHTmpl = fmt.Sprintf(
	"{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubeDockerdProfile }}{{ .Suffix }}"+
		"{{ .UsageHint }}",
	constants.DockerHostEnv,
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
	sshHost              bool
	sshAdd               bool
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
	if ec.ssh {
		usgCmd += " --ssh-host"
	}
	s := &DockerShellConfig{
		Config: *shell.CfgSet(ec.EnvConfig, usgPlz, usgCmd),
	}
	if !ec.ssh {
		s.DockerCertPath = envMap[constants.DockerCertPathEnv]
	}
	s.DockerHost = envMap[constants.DockerHostEnv]
	if !ec.ssh {
		s.DockerTLSVerify = envMap[constants.DockerTLSVerifyEnv]
	}

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
func ensureDockerd(name string, r command.Runner) {
	if ok := isDockerActive(r); ok {
		return
	}
	mustRestartDockerd(name, r)
}

// isDockerActive checks if Docker is active
func isDockerActive(r command.Runner) bool {
	return sysinit.New(r).Active("docker")
}

// mustRestartDockerd will attempt to reload dockerd if fails, will try restart and exit if fails again
func mustRestartDockerd(name string, runner command.Runner) {
	// Docker Docs: https://docs.docker.com/config/containers/live-restore
	// On Linux, you can avoid a restart (and avoid any downtime for your containers) by reloading the Docker daemon.
	klog.Warningf("dockerd is not active will try to reload it...")
	if err := sysinit.New(runner).Reload("docker"); err != nil {
		klog.Warningf("will try to restart dockerd because reload failed: %v", err)
		if err := sysinit.New(runner).Restart("docker"); err != nil {
			exit.Message(reason.RuntimeRestart, `The Docker service within '{{.name}}' is not active`, out.V{"name": name})
		}
		// if we get to the point that we have to restart docker (instead of reload)
		// will need to wait for apisever container to come up, this usually takes 5 seconds
		// verifying apisever using kverify would add code complexity for a rare case.
		klog.Warningf("waiting to ensure apisever container is up...")
		startTime := time.Now()

		if err = WaitForAPIServerProcess(runner, startTime, time.Second*5); err != nil {
			exit.Message(reason.RuntimeRestart, `The api server within '{{.name}}' is not up`, out.V{"name": name})
		}
	}
}

func WaitForAPIServerProcess(cr command.Runner, start time.Time, timeout time.Duration) error {
	klog.Infof("waiting for apiserver process to appear ...")
	err := apiWait.PollImmediate(time.Millisecond*500, timeout, func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during process check")
		}

		if time.Since(start) > minLogCheckTime {
			klog.Infof("waiting for apiserver process to appear ...")
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		if _, ierr := kverify.APIServerPID(cr); ierr != nil {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("apiserver process never appeared")
	}
	klog.Infof("duration metric: took %s to wait for apiserver process to appear ...", time.Since(start))
	return nil
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

		r := co.CP.Runner
		ensureDockerd(cname, r)

		d := co.CP.Host.Driver
		port := constants.DockerDaemonPort
		if driver.NeedsPortForward(driverName) {
			port, err = oci.ForwardedPort(driverName, cname, port)
			if err != nil {
				exit.Message(reason.DrvPortForward, "Error getting port binding for '{{.driver_name}} driver: {{.error}}", out.V{"driver_name": driverName, "error": err})
			}
		}

		hostname, err := d.GetSSHHostname()
		if err != nil {
			exit.Error(reason.IfSSHClient, "Error getting ssh client", err)
		}

		sshport, err := d.GetSSHPort()
		if err != nil {
			exit.Error(reason.IfSSHClient, "Error getting ssh client", err)
		}

		hostIP := co.CP.IP.String()
		ec := DockerEnvConfig{
			EnvConfig: sh,
			profile:   cname,
			driver:    driverName,
			ssh:       sshHost,
			hostIP:    hostIP,
			port:      port,
			certsDir:  localpath.MakeMiniPath("certs"),
			noProxy:   noProxy,
			username:  d.GetSSHUsername(),
			hostname:  hostname,
			sshport:   sshport,
			keypath:   d.GetSSHKeyPath(),
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
				// even though docker maybe running just fine it could be holding on to old certs and needs a refresh
				klog.Warningf("couldn't connect to docker inside minikube.  output: %s error: %v", string(out), err)
				mustRestartDockerd(cname, co.CP.Runner)
			}
		}

		if err := dockerSetScript(ec, os.Stdout); err != nil {
			exit.Error(reason.InternalDockerScript, "Error generating set output", err)
		}

		if sshAdd {
			klog.Infof("Adding %v", d.GetSSHKeyPath())

			path, err := exec.LookPath("ssh-add")
			if err != nil {
				exit.Error(reason.IfSSHClient, "Error with ssh-add", err)
			}

			cmd := exec.Command(path, d.GetSSHKeyPath())
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				exit.Error(reason.IfSSHClient, "Error with ssh-add", err)
			}
		}
	},
}

// DockerEnvConfig encapsulates all external inputs into shell generation for Docker
type DockerEnvConfig struct {
	shell.EnvConfig
	profile  string
	driver   string
	ssh      bool
	hostIP   string
	port     int
	certsDir string
	noProxy  bool
	username string
	hostname string
	sshport  int
	keypath  string
}

// dockerSetScript writes out a shell-compatible 'docker-env' script
func dockerSetScript(ec DockerEnvConfig, w io.Writer) error {
	var dockerSetEnvTmpl string
	if ec.ssh {
		dockerSetEnvTmpl = dockerEnvSSHTmpl
	} else {
		dockerSetEnvTmpl = dockerEnvTCPTmpl
	}
	envVars := dockerEnvVars(ec)
	return shell.SetScript(ec.EnvConfig, w, dockerSetEnvTmpl, dockerShellCfgSet(ec, envVars))
}

// dockerSetScript writes out a shell-compatible 'docker-env unset' script
func dockerUnsetScript(ec DockerEnvConfig, w io.Writer) error {
	vars := dockerEnvNames(ec)
	return shell.UnsetScript(ec.EnvConfig, w, vars)
}

// dockerURL returns a the docker endpoint URL for an ip/port pair.
func dockerURL(ip string, port int) string {
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(port)))
}

// sshURL returns the docker endpoint URL when using socket over ssh.
func sshURL(username string, hostname string, port int) string {
	// assumes standard /var/run/docker.sock as the path (not possible to set it at the moment)
	return fmt.Sprintf("ssh://%s@%s", username, net.JoinHostPort(hostname, strconv.Itoa(port)))
}

// dockerEnvVars gets the necessary docker env variables to allow the use of minikube's docker daemon
func dockerEnvVars(ec DockerEnvConfig) map[string]string {
	envTCP := map[string]string{
		constants.DockerTLSVerifyEnv:       "1",
		constants.DockerHostEnv:            dockerURL(ec.hostIP, ec.port),
		constants.DockerCertPathEnv:        ec.certsDir,
		constants.MinikubeActiveDockerdEnv: ec.profile,
	}
	envSSH := map[string]string{
		constants.DockerHostEnv:            sshURL(ec.username, ec.hostname, ec.sshport),
		constants.MinikubeActiveDockerdEnv: ec.profile,
	}

	var rt map[string]string
	if ec.ssh {
		rt = envSSH
	} else {
		rt = envTCP
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

// dockerEnvNames gets the necessary docker env variables to reset after using minikube's docker daemon
func dockerEnvNames(ec DockerEnvConfig) []string {
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
	return vars
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
	dockerEnvCmd.Flags().BoolVar(&sshHost, "ssh-host", false, "Use SSH connection instead of HTTPS (port 2376)")
	dockerEnvCmd.Flags().BoolVar(&sshAdd, "ssh-add", false, "Add SSH identity key to SSH authentication agent")
	dockerEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	dockerEnvCmd.Flags().BoolVarP(&dockerUnset, "unset", "u", false, "Unset variables instead of setting them")
}
