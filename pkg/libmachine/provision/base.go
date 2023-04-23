package provision

import (
	"bytes"
	"fmt"
	"text/template"

	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cruntime"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/runner"
)

type GenericSSHCommander struct {
	Driver drivers.Driver
}

type BaseProvisioner struct {
	OsReleaseID         string
	CRuntimeOptionsDir  string
	CRuntimeOptionsFile string
	Packages            []string
	OsReleaseInfo       *OsRelease
	Driver              drivers.Driver
	AuthOptions         auth.Options
	EngineOptions       cruntime.Options
	Runner              runner.Runner
}

// RunCommand runs a command inside the linux machine,
// using our reference to the driver as our channel
func (bp *BaseProvisioner) RunCommand(args string) (string, error) {
	return bp.Runner.RunCommand(args)
}

// Hostname gets the hostname of the linux machine
func (bp *BaseProvisioner) Hostname() (string, error) {
	return bp.RunCommand("hostname")
}

// SetHostname sets the hostname of the linux machine
func (bp *BaseProvisioner) SetHostname(hostname string) error {
	if _, err := bp.RunCommand(fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	// ubuntu/debian use 127.0.1.1 for non "localhost" loopback hostnames: https://www.debian.org/doc/manuals/debian-reference/ch05.en.html#_the_hostname_resolution
	if _, err := bp.RunCommand(fmt.Sprintf(`
		if ! grep -xq '.*\s%s' /etc/hosts; then
			if grep -xq '127.0.1.1\s.*' /etc/hosts; then
				sudo sed -i 's/^127.0.1.1\s.*/127.0.1.1 %s/g' /etc/hosts;
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

// GetDockerOptionsDir returns a path to the docker container runtime config dir
func (provisioner *BaseProvisioner) GetDockerOptionsDir() string {
	return provisioner.CRuntimeOptionsDir
}

func (provisioner *BaseProvisioner) CompatibleWithMachine() bool {
	return provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID
}

func (provisioner *BaseProvisioner) GetAuthOptions() auth.Options {
	return provisioner.AuthOptions
}

func (bp *BaseProvisioner) SetOsReleaseInfo(info *OsRelease) {
	bp.OsReleaseInfo = info
}

// GetOsReleaseInfo returns a struct referring to the /etc/os-release file
// inside the linux machine
func (bp *BaseProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return bp.OsReleaseInfo, nil
}

// TODO:
func (bp *BaseProvisioner) GenerateContainerRuntimeOptions(interface{}) (*ContainerRuntimeOptions, error) {
	return nil, nil
}

// GenerateDockerOptions generates an options struct that can be used to configure a docker daemon
func (bp *BaseProvisioner) GenerateDockerOptions(dockerPort int) (*ContainerRuntimeOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", bp.Driver.DriverName())
	bp.EngineOptions.Labels = append(bp.EngineOptions.Labels, driverNameLabel)

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
		AuthOptions:   bp.AuthOptions,
		EngineOptions: bp.EngineOptions,
	}

	t.Execute(&engineCfg, engineConfigContext)

	return &ContainerRuntimeOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: bp.CRuntimeOptionsFile,
	}, nil
}

// GetDriver returns a reference for the driver that we're making use of
func (provisioner *BaseProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
