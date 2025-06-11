package environment

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/drivers/qemu"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/shell"
	"k8s.io/minikube/pkg/minikube/sysinit"
	pkgnetwork "k8s.io/minikube/pkg/network"
)

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
	constants.MinikubeActiveDockerdEnv, // Using unified constant
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
	constants.MinikubeActiveDockerdEnv, // Using unified constant
	constants.SSHAuthSock,
	constants.SSHAgentPID)

// DockerConfigurator implements the EnvConfigurator interface for Docker and Containerd runtimes.
type DockerConfigurator struct {
	profile     string
	useSSH      bool
	noProxy     bool
	hostIP      string
	port        int
	certsDir    string
	username    string
	hostname    string
	sshport     int
	keypath     string
	sshAuthSock string
	sshAgentPID int
}

// NewDockerConfigurator creates a new Docker environment configurator.
func NewDockerConfigurator(co *mustload.ClusterController, useSSH bool, noProxy bool) (*DockerConfigurator, error) {
	cr := co.Config.KubernetesConfig.ContainerRuntime
	cname := co.Config.Name

	if cr == constants.Containerd {
		useSSH = true
	}

	if cr == constants.Docker {
		if err := ensureDockerd(cname, co.CP.Runner); err != nil {
			return nil, fmt.Errorf("dockerd not running: %w", err)
		}
	}

	d := co.CP.Host.Driver
	driverName := d.DriverName()
	port := constants.DockerDaemonPort
	var err error

	if driver.NeedsPortForward(driverName) {
		port, err = oci.ForwardedPort(driverName, co.Config.Name, port)
		if err != nil {
			return nil, fmt.Errorf("getting forwarded port: %w", err)
		}
	} else if driver.IsQEMU(driverName) && pkgnetwork.IsBuiltinQEMU(co.Config.Network) {
		port = d.(*qemu.Driver).EnginePort
	}

	hostname, err := d.GetSSHHostname()
	if err != nil {
		return nil, fmt.Errorf("getting ssh hostname: %w", err)
	}
	sshport, err := d.GetSSHPort()
	if err != nil {
		return nil, fmt.Errorf("getting ssh port: %w", err)
	}

	return &DockerConfigurator{
		profile:     co.Config.Name,
		useSSH:      useSSH,
		noProxy:     noProxy,
		hostIP:      co.CP.IP.String(),
		port:        port,
		certsDir:    localpath.MakeMiniPath("certs"),
		username:    d.GetSSHUsername(),
		hostname:    hostname,
		sshport:     sshport,
		keypath:     d.GetSSHKeyPath(),
		sshAuthSock: co.Config.SSHAuthSock,
		sshAgentPID: co.Config.SSHAgentPID,
	}, nil
}

// Vars implements the EnvConfigurator interface.
func (d *DockerConfigurator) Vars() (map[string]string, error) {
	agentPID := strconv.Itoa(d.sshAgentPID)
	if agentPID == "0" {
		agentPID = ""
	}

	env := make(map[string]string)

	if d.useSSH {
		env[constants.DockerHostEnv] = sshURL(d.username, d.hostname, d.sshport)
	} else {
		env[constants.DockerTLSVerifyEnv] = "1"
		env[constants.DockerHostEnv] = dockerURL(d.hostIP, d.port)
		env[constants.DockerCertPathEnv] = d.certsDir
	}

	env[constants.MinikubeActiveDockerdEnv] = d.profile
	if d.sshAuthSock != "" {
		env[constants.SSHAuthSock] = d.sshAuthSock
	}
	if agentPID != "" {
		env[constants.SSHAgentPID] = agentPID
	}

	if d.noProxy {
		noProxyVar, noProxyValue := GetNoProxyVar()
		if noProxyValue == "" {
			noProxyValue = d.hostIP
		} else if !strings.Contains(noProxyValue, d.hostIP) {
			noProxyValue = fmt.Sprintf("%s,%s", noProxyValue, d.hostIP)
		}
		env[noProxyVar] = noProxyValue
	}

	if os.Getenv(constants.MinikubeActiveDockerdEnv) == "" {
		for _, e := range constants.DockerDaemonEnvs {
			if v := oci.InitialEnv(e); v != "" {
				env[constants.MinikubeExistingPrefix+e] = v
			}
		}
	}
	return env, nil
}

// UnsetVars implements the EnvConfigurator interface.
func (d *DockerConfigurator) UnsetVars() ([]string, error) {
	vars := []string{
		constants.DockerTLSVerifyEnv,
		constants.DockerHostEnv,
		constants.DockerCertPathEnv,
		constants.MinikubeActiveDockerdEnv,
		constants.SSHAuthSock,
		constants.SSHAgentPID,
	}

	if d.noProxy {
		k, _ := GetNoProxyVar()
		if k != "" {
			vars = append(vars, k)
		}
	}
	return vars, nil
}

// DisplayScript implements the EnvConfigurator interface.
func (d *DockerConfigurator) DisplayScript(sh shell.Config, w io.Writer) error {
	vars, err := d.Vars()
	if err != nil {
		return err
	}
	cfg := d.createShellConfig(vars, sh)
	tmpl := dockerEnvTCPTmpl
	if d.useSSH {
		tmpl = dockerEnvSSHTmpl
	}
	return shell.SetScript(w, tmpl, cfg)
}

// DockerShellConfig represents the shell configuration of Docker 
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
	SSHAuthSock            string
	SSHAgentPID            string
}

func (d *DockerConfigurator) createShellConfig(envMap map[string]string, sh shell.Config) *DockerShellConfig {
	const usgPlz = "To point your shell to minikube's container engine, run:"
	usgCmd := fmt.Sprintf("minikube -p %s docker-env", d.profile)
	if d.useSSH {
		usgCmd += " --ssh-host"
	}
	s := &DockerShellConfig{Config: *shell.CfgSet(sh, usgPlz, usgCmd)}
	s.DockerHost = envMap[constants.DockerHostEnv]
	s.MinikubeDockerdProfile = envMap[constants.MinikubeActiveDockerdEnv]
	s.SSHAuthSock = envMap[constants.SSHAuthSock]
	s.SSHAgentPID = envMap[constants.SSHAgentPID]
	s.ExistingDockerHost = envMap[constants.ExistingDockerHostEnv]
	s.ExistingDockerCertPath = envMap[constants.ExistingDockerCertPathEnv]
	s.ExistingDockerTLSVerify = envMap[constants.ExistingDockerTLSVerifyEnv]
	if !d.useSSH {
		s.DockerCertPath = envMap[constants.DockerCertPathEnv]
		s.DockerTLSVerify = envMap[constants.DockerTLSVerifyEnv]
	}
	if d.noProxy {
		noProxyVar, _ := GetNoProxyVar()
		s.NoProxyVar = noProxyVar
		s.NoProxyValue = envMap[noProxyVar]
	}
	return s
}

func ensureDockerd(name string, r command.Runner) error {
	if sysinit.New(r).Active("docker") {
		return nil
	}
	if err := sysinit.New(r).Reload("docker"); err != nil {
		if err := sysinit.New(r).Restart("docker"); err != nil {
			return fmt.Errorf("couldn't restart docker inside minikube within '%s': %w", name, err)
		}
	}
	return nil
}

func dockerURL(ip string, port int) string {
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(port)))
}
func sshURL(username string, hostname string, port int) string {
	return fmt.Sprintf("ssh://%s@%s", username, net.JoinHostPort(hostname, strconv.Itoa(port)))
}