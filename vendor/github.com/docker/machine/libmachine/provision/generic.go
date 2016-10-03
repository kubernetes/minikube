package provision

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
)

type GenericProvisioner struct {
	SSHCommander
	OsReleaseID       string
	DockerOptionsDir  string
	DaemonOptionsFile string
	Packages          []string
	OsReleaseInfo     *OsRelease
	Driver            drivers.Driver
	AuthOptions       auth.Options
	EngineOptions     engine.Options
	SwarmOptions      swarm.Options
}

type GenericSSHCommander struct {
	Driver drivers.Driver
}

func (sshCmder GenericSSHCommander) SSHCommand(args string) (string, error) {
	return drivers.RunSSHCommandFromDriver(sshCmder.Driver, args)
}

func (provisioner *GenericProvisioner) Hostname() (string, error) {
	return provisioner.SSHCommand("hostname")
}

func (provisioner *GenericProvisioner) SetHostname(hostname string) error {
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	// ubuntu/debian use 127.0.1.1 for non "localhost" loopback hostnames: https://www.debian.org/doc/manuals/debian-reference/ch05.en.html#_the_hostname_resolution
	if _, err := provisioner.SSHCommand(fmt.Sprintf(`
		if ! grep -xq .*%s /etc/hosts; then
			if grep -xq 127.0.1.1.* /etc/hosts; then 
				sudo sed -i 's/^127.0.1.1.*/127.0.1.1 %s/g' /etc/hosts; 
			else 
				echo '127.0.1.1 %s' | sudo tee -a /etc/hosts; 
			fi
		fi`,
		hostname,
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	return nil
}

func (provisioner *GenericProvisioner) GetDockerOptionsDir() string {
	return provisioner.DockerOptionsDir
}

func (provisioner *GenericProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID
}

func (provisioner *GenericProvisioner) GetAuthOptions() auth.Options {
	return provisioner.AuthOptions
}

func (provisioner *GenericProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}

func (provisioner *GenericProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return provisioner.OsReleaseInfo, nil
}

func (provisioner *GenericProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `
DOCKER_OPTS='
-H tcp://0.0.0.0:{{.DockerPort}}
-H unix:///var/run/docker.sock
--storage-driver {{.EngineOptions.StorageDriver}}
--tlsverify
--tlscacert {{.AuthOptions.CaCertRemotePath}}
--tlscert {{.AuthOptions.ServerCertRemotePath}}
--tlskey {{.AuthOptions.ServerKeyRemotePath}}
{{ range .EngineOptions.Labels }}--label {{.}}
{{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}}
{{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}}
{{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}}
{{ end }}
'
{{range .EngineOptions.Env}}export \"{{ printf "%q" . }}\"
{{end}}
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		DockerPort:    dockerPort,
		AuthOptions:   provisioner.AuthOptions,
		EngineOptions: provisioner.EngineOptions,
	}

	t.Execute(&engineCfg, engineConfigContext)

	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: provisioner.DaemonOptionsFile,
	}, nil
}

func (provisioner *GenericProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
