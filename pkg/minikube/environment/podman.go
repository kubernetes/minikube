package environment

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/ssh"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/shell"
)

var podmanEnv1Tmpl = fmt.Sprintf(
	"{{ .Prefix }}%s{{ .Delimiter }}{{ .VarlinkBridge }}{{ .Suffix }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubePodmanProfile }}{{ .Suffix }}"+
		"{{ .UsageHint }}",
	constants.PodmanVarlinkBridgeEnv,
	constants.MinikubeActivePodmanEnv) // Using unified constant

var podmanEnv2Tmpl = fmt.Sprintf(
	"{{ .Prefix }}%s{{ .Delimiter }}{{ .ContainerHost }}{{ .Suffix }}"+
		"{{ if .ContainerSSHKey }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .ContainerSSHKey}}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ if .ExistingContainerHost }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .ExistingContainerHost }}{{ .Suffix }}"+
		"{{ end }}"+
		"{{ .Prefix }}%s{{ .Delimiter }}{{ .MinikubePodmanProfile }}{{ .Suffix }}"+
		"{{ .UsageHint }}",
	constants.PodmanContainerHostEnv,
	constants.PodmanContainerSSHKeyEnv,
	constants.ExistingContainerHostEnv,
	constants.MinikubeActivePodmanEnv) // Using unified constant

// PodmanConfigurator 为 Podman 实现 EnvConfigurator 接口。
type PodmanConfigurator struct {
	profile  string
	varlink  bool
	client   *ssh.ExternalClient
	username string
	hostname string
	port     int
	keypath  string
}

// NewPodmanConfigurator 创建一个新的 Podman 环境配置器。
func NewPodmanConfigurator(co *mustload.ClusterController) (*PodmanConfigurator, error) {
	r := co.CP.Runner
	varlink, err := isVarlinkAvailable(r)
	if err != nil {
		return nil, fmt.Errorf("failed to check for varlink: %w", err)
	}
	d := co.CP.Host.Driver
	client, err := createExternalSSHClient(d)
	if err != nil {
		return nil, fmt.Errorf("error getting ssh client: %w", err)
	}
	hostname, err := d.GetSSHHostname()
	if err != nil {
		return nil, fmt.Errorf("error getting ssh hostname: %w", err)
	}
	port, err := d.GetSSHPort()
	if err != nil {
		return nil, fmt.Errorf("error getting ssh port: %w", err)
	}
	return &PodmanConfigurator{
		profile:  co.Config.Name,
		varlink:  varlink,
		client:   client,
		username: d.GetSSHUsername(),
		hostname: hostname,
		port:     port,
		keypath:  d.GetSSHKeyPath(),
	}, nil
}

// Vars 实现了 EnvConfigurator 接口。
func (p *PodmanConfigurator) Vars() (map[string]string, error) {
	var env map[string]string
	if p.varlink {
		env = map[string]string{
			constants.PodmanVarlinkBridgeEnv: podmanBridge(p.client),
		}
	} else {
		env = map[string]string{
			constants.PodmanContainerHostEnv:   podmanSSHURL(p.username, p.hostname, p.port),
			constants.PodmanContainerSSHKeyEnv: p.keypath,
		}
	}

	env[constants.MinikubeActivePodmanEnv] = p.profile
	if os.Getenv(constants.MinikubeActivePodmanEnv) == "" {
		if v := oci.InitialEnv(constants.PodmanContainerHostEnv); v != "" {
			env[constants.ExistingContainerHostEnv] = v
		}
	}
	return env, nil
}

// UnsetVars 实现了 EnvConfigurator 接口。
func (p *PodmanConfigurator) UnsetVars() ([]string, error) {
	vars := []string{constants.MinikubeActivePodmanEnv}
	if p.varlink {
		vars = append(vars, constants.PodmanVarlinkBridgeEnv)
	} else {
		vars = append(vars, constants.PodmanContainerHostEnv, constants.PodmanContainerSSHKeyEnv)
	}
	return vars, nil
}

// DisplayScript 实现了 EnvConfigurator 接口。
func (p *PodmanConfigurator) DisplayScript(sh shell.Config, w io.Writer) error {
	vars, err := p.Vars()
	if err != nil {
		return err
	}
	cfg := p.createShellConfig(vars, sh)
	tmpl := podmanEnv2Tmpl
	if p.varlink {
		tmpl = podmanEnv1Tmpl
	}
	return shell.SetScript(w, tmpl, cfg)
}

// --- Podman 的辅助函数 ---
type PodmanShellConfig struct {
	shell.Config
	VarlinkBridge         string
	ContainerHost         string
	ContainerSSHKey       string
	MinikubePodmanProfile string
	ExistingContainerHost string
}

func (p *PodmanConfigurator) createShellConfig(envMap map[string]string, sh shell.Config) *PodmanShellConfig {
	const usgPlz = "To point your shell to minikube's container engine, run:"
	usgCmd := fmt.Sprintf("minikube -p %s docker-env", p.profile)
	return &PodmanShellConfig{
		Config:                *shell.CfgSet(sh, usgPlz, usgCmd),
		VarlinkBridge:         envMap[constants.PodmanVarlinkBridgeEnv],
		ContainerHost:         envMap[constants.PodmanContainerHostEnv],
		ContainerSSHKey:       envMap[constants.PodmanContainerSSHKeyEnv],
		MinikubePodmanProfile: envMap[constants.MinikubeActivePodmanEnv],
		ExistingContainerHost: envMap[constants.ExistingContainerHostEnv],
	}
}
func isVarlinkAvailable(r command.Runner) (bool, error) {
	_, err := r.RunCmd(exec.Command("which", "varlink"))
	return err == nil, nil
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
func podmanBridge(client *ssh.ExternalClient) string {
	cmd := []string{client.BinaryPath}
	cmd = append(cmd, client.BaseArgs...)
	cmd = append(cmd, "--", "sudo", "varlink", "-A", `\'podman varlink \\\$VARLINK_ADDRESS\'`, "bridge")
	return strings.Join(cmd, " ")
}
func podmanSSHURL(username string, hostname string, port int) string {
	path := "/run/podman/podman.sock"
	return fmt.Sprintf("ssh://%s@%s:%d%s", username, hostname, port, path)
}