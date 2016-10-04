package provision

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/provision/serviceaction"
)

type SystemdProvisioner struct {
	GenericProvisioner
}

func (p *SystemdProvisioner) String() string {
	return "redhat"
}

func NewSystemdProvisioner(osReleaseID string, d drivers.Driver) SystemdProvisioner {
	return SystemdProvisioner{
		GenericProvisioner{
			SSHCommander:      GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/systemd/system/docker.service",
			OsReleaseID:       osReleaseID,
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

func (p *SystemdProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", p.Driver.DriverName())
	p.EngineOptions.Labels = append(p.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `[Service]
ExecStart=/usr/bin/docker daemon -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock --storage-driver {{.EngineOptions.StorageDriver}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
MountFlags=slave
LimitNOFILE=1048576
LimitNPROC=1048576
LimitCORE=infinity
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}

[Install]
WantedBy=multi-user.target
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

	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: p.DaemonOptionsFile,
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
		if _, err := p.SSHCommand("sudo systemctl daemon-reload"); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("sudo systemctl -f %s %s", action.String(), name)

	if _, err := p.SSHCommand(command); err != nil {
		return err
	}

	return nil
}
