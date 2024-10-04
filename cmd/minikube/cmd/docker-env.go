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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	apiWait "k8s.io/apimachinery/pkg/util/wait"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/drivers/qemu"
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
	"k8s.io/minikube/pkg/minikube/sshagent"
	"k8s.io/minikube/pkg/minikube/sysinit"
	pkgnetwork "k8s.io/minikube/pkg/network"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
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
		"{{ if .SSHAuthSock }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .SSHAuthSock }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ if .SSHAgentPID }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .SSHAgentPID }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ .UsageHint }}",
	constants.DockerTLSVerifyEnv,
	constants.DockerHostEnv,
	constants.DockerCertPathEnv,
	constants.ExistingDockerTLSVerifyEnv,
	constants.ExistingDockerHostEnv,
	constants.ExistingDockerCertPathEnv,
	constants.MinikubeActiveDockerdEnv,
	constants.SSHAuthSock,
	constants.SSHAgentPID)
var dockerEnvSSHTmpl = fmt.Sprintf(
	"{{ .Prefix }}%s{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubeDockerdProfile }}{{ .Suffix }}"+
		"{{ if .SSHAuthSock }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .SSHAuthSock }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ if .SSHAgentPID }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .SSHAgentPID }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ .UsageHint }}",
	constants.DockerHostEnv,
	constants.MinikubeActiveDockerdEnv,
	constants.SSHAuthSock,
	constants.SSHAgentPID)

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

	SSHAuthSock string
	SSHAgentPID string
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

	s.SSHAuthSock = envMap[constants.SSHAuthSock]
	s.SSHAgentPID = envMap[constants.SSHAgentPID]

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
			klog.Warningf("Couldn't restart docker inside minikube within '%v' because: %v", name, err)
			return
		}
		// if we get to the point that we have to restart docker (instead of reload)
		// will need to wait for apisever container to come up, this usually takes 5 seconds
		// verifying apisever using kverify would add code complexity for a rare case.
		klog.Warningf("waiting to ensure apisever container is up...")
		if err = waitForAPIServerProcess(runner, time.Now(), time.Second*30); err != nil {
			klog.Warningf("apiserver container isn't up, error: %v", err)
		}
	}
}

func waitForAPIServerProcess(cr command.Runner, start time.Time, timeout time.Duration) error {
	klog.Infof("waiting for apiserver process to appear ...")
	err := apiWait.PollUntilContextTimeout(context.Background(), time.Millisecond*500, timeout, true, func(_ context.Context) (bool, error) {
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
	Short: "Provides instructions to point your terminal's docker-cli to the Docker Engine inside minikube. (Useful for building docker images directly inside minikube)",
	Long: `Provides instructions to point your terminal's docker-cli to the Docker Engine inside minikube. (Useful for building docker images directly inside minikube)

For example, you can do all docker operations such as docker build, docker run, and docker ps directly on the docker inside minikube.

Note: You need the docker-cli to be installed on your machine.
docker-cli install instructions: https://minikube.sigs.k8s.io/docs/tutorials/docker_desktop_replacement/#steps`,
	Run: func(_ *cobra.Command, _ []string) {
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

		if !out.IsTerminal(os.Stdout) {
			out.SetSilent(true)
			exit.SetShell(true)
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
		cr := co.Config.KubernetesConfig.ContainerRuntime
		if err := dockerEnvSupported(cr, driverName); err != nil {
			exit.Message(reason.Usage, err.Error())
		}

		// for the sake of docker-env command, start nerdctl and nerdctld
		if cr == constants.Containerd {
			out.WarningT("Using the docker-env command with the containerd runtime is a highly experimental feature, please provide feedback or contribute to make it better")

			startNerdctld()

			// docker-env on containerd depends on nerdctld (https://github.com/afbjorklund/nerdctld) as "docker" daeomn
			// and nerdctld daemon must be used with ssh connection (it is set in kicbase image's Dockerfile)
			// so directly set --ssh-host --ssh-add to true, even user didn't specify them
			sshAdd = true
			sshHost = true

			// start the ssh-agent
			if err := sshagent.Start(cname); err != nil {
				exit.Message(reason.SSHAgentStart, err.Error())
			}
			// cluster config must be reloaded
			// otherwise we won't be able to get SSH_AUTH_SOCK and SSH_AGENT_PID from cluster config.
			co = mustload.Running(cname)

			// set the ssh-agent envs for current process
			os.Setenv("SSH_AUTH_SOCK", co.Config.SSHAuthSock)
			os.Setenv("SSH_AGENT_PID", strconv.Itoa(co.Config.SSHAgentPID))
		}

		r := co.CP.Runner

		if cr == constants.Docker {
			ensureDockerd(cname, r)
		}

		d := co.CP.Host.Driver
		port := constants.DockerDaemonPort
		if driver.NeedsPortForward(driverName) {
			port, err = oci.ForwardedPort(driverName, cname, port)
			if err != nil {
				exit.Message(reason.DrvPortForward, "Error getting port binding for '{{.driver_name}} driver: {{.error}}", out.V{"driver_name": driverName, "error": err})
			}
		} else if driver.IsQEMU(driverName) && pkgnetwork.IsBuiltinQEMU(co.Config.Network) {
			port = d.(*qemu.Driver).EnginePort
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
			EnvConfig:   sh,
			profile:     cname,
			driver:      driverName,
			ssh:         sshHost,
			hostIP:      hostIP,
			port:        port,
			certsDir:    localpath.MakeMiniPath("certs"),
			noProxy:     noProxy,
			username:    d.GetSSHUsername(),
			hostname:    hostname,
			sshport:     sshport,
			keypath:     d.GetSSHKeyPath(),
			sshAuthSock: co.Config.SSHAuthSock,
			sshAgentPID: co.Config.SSHAgentPID,
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

			// TODO: refactor to work with docker, temp fix to resolve regression
			if cr == constants.Containerd {
				cmd.Env = append(cmd.Env, fmt.Sprintf("SSH_AUTH_SOCK=%s", co.Config.SSHAuthSock))
				cmd.Env = append(cmd.Env, fmt.Sprintf("SSH_AGENT_PID=%d", co.Config.SSHAgentPID))
			}
			err = cmd.Run()
			if err != nil {
				exit.Error(reason.IfSSHClient, "Error with ssh-add", err)
			}

			// TODO: refactor to work with docker, temp fix to resolve regression
			if cr == constants.Containerd {
				// eventually, run something similar to ssh --append-known
				appendKnownHelper(nodeName, true)
			}
		}
	},
}

// DockerEnvConfig encapsulates all external inputs into shell generation for Docker
type DockerEnvConfig struct {
	shell.EnvConfig
	profile     string
	driver      string
	ssh         bool
	hostIP      string
	port        int
	certsDir    string
	noProxy     bool
	username    string
	hostname    string
	sshport     int
	keypath     string
	sshAuthSock string
	sshAgentPID int
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
	if ec.Shell == "none" {
		switch outputFormat {
		case "":
			// shell "none"
			break
		case "text":
			for k, v := range envVars {
				_, err := fmt.Fprintf(w, "%s=%s\n", k, v)
				if err != nil {
					return err
				}
			}
			return nil
		case "json":
			json, err := json.Marshal(envVars)
			if err != nil {
				return err
			}
			_, err = w.Write(json)
			if err != nil {
				return err
			}
			_, err = w.Write([]byte{'\n'})
			if err != nil {
				return err
			}
			return nil
		case "yaml":
			yaml, err := yaml.Marshal(envVars)
			if err != nil {
				return err
			}
			_, err = w.Write(yaml)
			if err != nil {
				return err
			}
			return nil
		default:
			exit.Message(reason.InternalOutputUsage, "error: --output must be 'text', 'yaml' or 'json'")
		}
	}
	return shell.SetScript(w, dockerSetEnvTmpl, dockerShellCfgSet(ec, envVars))
}

// dockerUnsetScript writes out a shell-compatible 'docker-env unset' script
func dockerUnsetScript(ec DockerEnvConfig, w io.Writer) error {
	vars := dockerEnvNames(ec)
	if ec.Shell == "none" {
		switch outputFormat {
		case "":
			// shell "none"
			break
		case "text":
			for _, n := range vars {
				_, err := fmt.Fprintf(w, "%s\n", n)
				if err != nil {
					return err
				}
			}
			return nil
		case "json":
			json, err := json.Marshal(vars)
			if err != nil {
				return err
			}
			_, err = w.Write(json)
			if err != nil {
				return err
			}
			_, err = w.Write([]byte{'\n'})
			if err != nil {
				return err
			}
			return nil
		case "yaml":
			yaml, err := yaml.Marshal(vars)
			if err != nil {
				return err
			}
			_, err = w.Write(yaml)
			if err != nil {
				return err
			}
			return nil
		default:
			exit.Message(reason.InternalOutputUsage, "error: --output must be 'text', 'yaml' or 'json'")
		}
	}
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
	agentPID := strconv.Itoa(ec.sshAgentPID)
	// set agentPID to nil value if not set
	if agentPID == "0" {
		agentPID = ""
	}
	envTCP := map[string]string{
		constants.DockerTLSVerifyEnv:       "1",
		constants.DockerHostEnv:            dockerURL(ec.hostIP, ec.port),
		constants.DockerCertPathEnv:        ec.certsDir,
		constants.MinikubeActiveDockerdEnv: ec.profile,
		constants.SSHAuthSock:              ec.sshAuthSock,
		constants.SSHAgentPID:              agentPID,
	}
	envSSH := map[string]string{
		constants.DockerHostEnv:            sshURL(ec.username, ec.hostname, ec.sshport),
		constants.MinikubeActiveDockerdEnv: ec.profile,
		constants.SSHAuthSock:              ec.sshAuthSock,
		constants.SSHAgentPID:              agentPID,
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
		constants.SSHAuthSock,
		constants.SSHAgentPID,
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
		fmt.Sprintf("%s=%s", constants.SSHAuthSock, ec.sshAuthSock),
		fmt.Sprintf("%s=%d", constants.SSHAgentPID, ec.sshAgentPID),
	}
}

func isValidDockerProxy(env string) bool {
	val := os.Getenv(env)
	if val == "" {
		return true
	}

	u, err := url.Parse(val)
	if err != nil {
		klog.Warningf("Parsing proxy env variable %s=%s error: %v", env, val, err)
		return false
	}
	switch u.Scheme {
	// See moby/moby#25740
	case "socks5", "socks5h":
		return true
	default:
		return false
	}
}

func removeInvalidDockerProxy() {
	for _, env := range []string{"ALL_PROXY", "all_proxy"} {
		if !isValidDockerProxy(env) {
			klog.Warningf("Ignoring non socks5 proxy env variable %s=%s", env, os.Getenv(env))
			os.Unsetenv(env)
		}
	}
}

// tryDockerConnectivity will try to connect to docker env from user's POV to detect the problem if it needs reset or not
func tryDockerConnectivity(bin string, ec DockerEnvConfig) ([]byte, error) {
	c := exec.Command(bin, "version", "--format={{.Server}}")

	// See #10098 for details
	removeInvalidDockerProxy()
	c.Env = append(os.Environ(), dockerEnvVarsList(ec)...)
	klog.Infof("Testing Docker connectivity with: %v", c)
	return c.CombinedOutput()
}

func dockerEnvSupported(containerRuntime, driverName string) error {
	if containerRuntime != constants.Docker && containerRuntime != constants.Containerd {
		return fmt.Errorf("the docker-env command only supports the docker and containerd runtimes")
	}
	// we only support containerd-env on the Docker driver
	if containerRuntime == constants.Containerd && driverName != driver.Docker {
		return fmt.Errorf("the docker-env command only supports the containerd runtime with the docker driver")
	}
	return nil
}

func init() {
	defaultNoProxyGetter = &EnvNoProxyGetter{}
	dockerEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add machine IP to NO_PROXY environment variable")
	dockerEnvCmd.Flags().BoolVar(&sshHost, "ssh-host", false, "Use SSH connection instead of HTTPS (port 2376)")
	dockerEnvCmd.Flags().BoolVar(&sshAdd, "ssh-add", false, "Add SSH identity key to SSH authentication agent")
	dockerEnvCmd.Flags().StringVar(&shell.ForceShell, "shell", "", "Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect")
	dockerEnvCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "One of 'text', 'yaml' or 'json'.")
	dockerEnvCmd.Flags().BoolVarP(&dockerUnset, "unset", "u", false, "Unset variables instead of setting them")
}
