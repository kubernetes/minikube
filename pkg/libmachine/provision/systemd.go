package provision

import (
	"bytes"
	"fmt"
	"os/exec"
	"text/template"

	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cruntime"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/provision/serviceaction"
	"k8s.io/minikube/pkg/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/versioncmp"
	"k8s.io/minikube/pkg/minikube/assets"
)

type SystemdProvisioner struct {
	BaseProvisioner
}

func (p *SystemdProvisioner) GetProvisionerName() string {
	return "redhat"
}

func NewSystemdProvisioner(osReleaseID string, d drivers.Driver) SystemdProvisioner {
	return SystemdProvisioner{
		BaseProvisioner{
			Runner:              d.GetRunner(),
			CRuntimeOptionsDir:  "/etc/docker",
			CRuntimeOptionsFile: "/etc/systemd/system/docker.service.d/10-machine.conf",
			OsReleaseID:         osReleaseID,
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

func (p *SystemdProvisioner) GenerateDockerOptions(dockerPort int) (*ContainerRuntimeOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", p.Driver.DriverName())
	p.EngineOptions.Labels = append(p.EngineOptions.Labels, driverNameLabel)

	dockerVersion, err := DockerClientVersion(p)
	if err != nil {
		return nil, err
	}

	arg := "dockerd"
	if versioncmp.LessThan(dockerVersion, "1.12.0") {
		arg = "docker daemon"
	}

	engineConfigTmpl := `[Service]
ExecStart=
ExecStart=/usr/bin/` + arg + ` -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock --storage-driver {{.EngineOptions.StorageDriver}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		DockerPort:    dockerPort,
		AuthOptions:   p.AuthOptions,
		EngineOptions: p.EngineOptions,
	}

	t.Execute(&engineCfg, engineConfigContext)

	return &ContainerRuntimeOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: p.CRuntimeOptionsFile,
	}, nil
}

func (p *SystemdProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	reloadDaemon := false
	switch action {
	case serviceaction.Start, serviceaction.Restart:
		reloadDaemon = true
	}

	// systemd needs reloaded when config changes on disk; we cannot
	// be sure exactly when it changes from the provisioner so
	// we call a reload on every restart to be safe
	if reloadDaemon {
		if _, err := p.Runner.RunCmd(exec.Command("sudo systemctl daemon-reload")); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("sudo systemctl -f %s %s", action.String(), name)

	if _, err := p.Runner.RunCmd(exec.Command(command)); err != nil {
		return err
	}

	return nil
}

// NOTE:
// obsiously too much for the systemd provisioner
// Need more refactoring.. this one should not be something that could
// be picked for provisioning..
func (p *SystemdProvisioner) Copy(assets.CopyableFile) error
func (p *SystemdProvisioner) CopyFrom(assets.CopyableFile) error
func (p *SystemdProvisioner) GetContainerRuntime() string
func (p *SystemdProvisioner) GetContainerRuntimeOptionsDir() string
func (p *SystemdProvisioner) GenerateContainerRuntimeOptions(interface{}) (*ContainerRuntimeOptions, error)
func (p *SystemdProvisioner) PackageAction(string, pkgaction.PackageAction) error
func (p *SystemdProvisioner) Provision(auth.Options, cruntime.Options) error
func (p *SystemdProvisioner) ReadableFile(sourcePath string) (assets.ReadableFile, error)
func (p *SystemdProvisioner) RemoveFile(assets.CopyableFile) error
func (p *SystemdProvisioner) RunCmd(cmd *exec.Cmd) (*runner.RunResult, error)
func (p *SystemdProvisioner) ServiceAction(string, serviceaction.ServiceAction) error
func (p *SystemdProvisioner) StartCmd(cmd *exec.Cmd) (*runner.StartedCmd, error)
func (p *SystemdProvisioner) WaitCmd(startedCmd *runner.StartedCmd) (*runner.RunResult, error)
