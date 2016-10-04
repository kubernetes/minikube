package provision

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/swarm"
)

func init() {
	Register("openSUSE", &RegisteredProvisioner{
		New: NewOpenSUSEProvisioner,
	})
	Register("SUSE Linux Enterprise Desktop", &RegisteredProvisioner{
		New: NewSLEDProvisioner,
	})
	Register("SUSE Linux Enterprise Server", &RegisteredProvisioner{
		New: NewSLESProvisioner,
	})
}

func NewOpenSUSEProvisioner(d drivers.Driver) Provisioner {
	return &SUSEProvisioner{
		GenericProvisioner{
			SSHCommander:      GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/sysconfig/docker",
			OsReleaseID:       "opensuse",
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

func NewSLEDProvisioner(d drivers.Driver) Provisioner {
	return &SUSEProvisioner{
		GenericProvisioner{
			SSHCommander:      GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/sysconfig/docker",
			OsReleaseID:       "sled",
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

func NewSLESProvisioner(d drivers.Driver) Provisioner {
	return &SUSEProvisioner{
		GenericProvisioner{
			SSHCommander:      GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/sysconfig/docker",
			OsReleaseID:       "sles",
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

type SUSEProvisioner struct {
	GenericProvisioner
}

func (provisioner *SUSEProvisioner) String() string {
	return "suse"
}

func (provisioner *SUSEProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	reloadDaemon := false
	switch action {
	case serviceaction.Start, serviceaction.Restart:
		reloadDaemon = true
	}

	// systemd needs reloaded when config changes on disk; we cannot
	// be sure exactly when it changes from the provisioner so
	// we call a reload on every restart to be safe
	if reloadDaemon {
		if _, err := provisioner.SSHCommand("sudo systemctl daemon-reload"); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("sudo systemctl %s %s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *SUSEProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	switch action {
	case pkgaction.Install:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
	case pkgaction.Upgrade:
		packageAction = "upgrade"
	}

	command := fmt.Sprintf("sudo -E zypper -n %s %s", packageAction, name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *SUSEProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *SUSEProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	// update OS -- this is needed for libdevicemapper and the docker install
	if _, err := provisioner.SSHCommand("sudo zypper ref"); err != nil {
		return err
	}
	if _, err := provisioner.SSHCommand("sudo zypper -n update"); err != nil {
		return err
	}

	if err := installDockerGeneric(provisioner, engineOptions.InstallURL); err != nil {
		return err
	}

	if _, err := provisioner.SSHCommand("sudo systemctl start docker"); err != nil {
		return err
	}

	if err := mcnutils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	if _, err := provisioner.SSHCommand("sudo systemctl stop docker"); err != nil {
		return err
	}

	// open firewall port required by docker
	if _, err := provisioner.SSHCommand("sudo /sbin/yast2 firewall services add ipprotocol=tcp tcpport=2376 zone=EXT"); err != nil {
		return err
	}

	if err := makeDockerOptionsDir(provisioner); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	if err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions); err != nil {
		return err
	}

	return nil
}

func (provisioner *SUSEProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg  bytes.Buffer
		configPath = provisioner.DaemonOptionsFile
	)

	// remove existing
	if _, err := provisioner.SSHCommand(fmt.Sprintf("sudo rm %s", configPath)); err != nil {
		return nil, err
	}

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `# File automatically generated by docker-machine
DOCKER_OPTS=' -H tcp://0.0.0.0:{{.DockerPort}} {{ if .EngineOptions.StorageDriver }} --storage-driver {{.EngineOptions.StorageDriver}} {{ end }} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}'
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		DockerPort:       dockerPort,
		AuthOptions:      provisioner.AuthOptions,
		EngineOptions:    provisioner.EngineOptions,
		DockerOptionsDir: provisioner.DockerOptionsDir,
	}

	t.Execute(&engineCfg, engineConfigContext)

	daemonOptsDir := configPath
	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: daemonOptsDir,
	}, nil
}
