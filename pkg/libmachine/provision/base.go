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
	"k8s.io/minikube/pkg/minikube/assets"
)

type BaseProvisioner struct {
	OsReleaseID         string
	CRuntime            string
	CRuntimeOptions     cruntime.Options
	CRuntimeOptionsDir  string
	CRuntimeOptionsFile string
	Packages            []string
	OsReleaseInfo       *OsRelease
	Driver              drivers.Driver
}

// GetProvisionerName returns the name of the provisioner we're using.
// This has to be overridden.
func (bp *BaseProvisioner) GetProvisionerName() string {
	return "BaseProvisioner"
}

// RunCommand runs a command inside the linux machine,
// using our reference to the driver as our channel
func (bp *BaseProvisioner) RunCmd(args string) (string, error) {
	rr, err := bp.Driver.GetRunner().RunCmd(exec.Command(args))
	return rr.Stdout.String(), err
}

// StartCmd starts a cmd of exec.Cmd type.
// This func in non-blocking, use WaitCmd to block until complete.
func (bp *BaseProvisioner) StartCmd(cmd *exec.Cmd) (*runner.StartedCmd, error) {
	return bp.Driver.GetRunner().StartCmd(cmd)
}

// WaitCmd will prevent further execution until the started command has completed
func (bp *BaseProvisioner) WaitCmd(startedCmd *runner.StartedCmd) (*runner.RunResult, error) {
	return bp.Driver.GetRunner().WaitCmd(startedCmd)
}

// Copy is a convenience method that runs a command to copy a file
func (bp *BaseProvisioner) Copy(cpblF assets.CopyableFile) error {
	return bp.Driver.GetRunner().Copy(cpblF)
}

// CopyFrom is a convenience method that runs a command to copy a file back
func (bp *BaseProvisioner) CopyFrom(cpblF assets.CopyableFile) error {
	return bp.Driver.GetRunner().CopyFrom(cpblF)
}

// Remove is a convenience method that runs a command to remove a file
func (bp *BaseProvisioner) RemoveFile(cpblF assets.CopyableFile) error {
	return bp.Driver.GetRunner().RemoveFile(cpblF)
}

// ReadableFile open a remote file for reading
func (bp *BaseProvisioner) ReadableFile(sourcePath string) (assets.ReadableFile, error) {
	return bp.Driver.GetRunner().ReadableFile(sourcePath)
}

// Hostname gets the hostname of the linux machine
func (bp *BaseProvisioner) Hostname() (string, error) {
	return bp.RunCmd("hostname")
}

// SetHostname sets the hostname of the linux machine
func (bp *BaseProvisioner) SetHostname(hostname string) error {
	if _, err := bp.RunCmd(fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	// ubuntu/debian use 127.0.1.1 for non "localhost" loopback hostnames: https://www.debian.org/doc/manuals/debian-reference/ch05.en.html#_the_hostname_resolution
	if _, err := bp.RunCmd(fmt.Sprintf(`
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

// GetCRuntimeOptionsDir returns a path to the container runtime config dir
func (bp *BaseProvisioner) GetCRuntimeOptionsDir() string {
	return bp.CRuntimeOptionsDir
}

// GetCRuntime returns the name of the container runtime that will run kubernetes
func (bp *BaseProvisioner) GetCRuntime() string {
	return bp.CRuntime
}

// GenerateCRuntimeOptions generates a CRuntimeOptsions struct
func (bp *BaseProvisioner) GenerateCRuntimeOptions(port int) (*ContainerRuntimeOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", bp.Driver.DriverName())
	bp.CRuntimeOptions.Labels = append(bp.CRuntimeOptions.Labels, driverNameLabel)

	engineConfigTmpl := `
DOCKER_OPTS='
-H tcp://0.0.0.0:{{.port}}
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

	engineConfigContext := CRuntimeConfigContext{
		Port:          port,
		AuthOptions:   bp.AuthOptions,
		EngineOptions: bp.CRuntimeOptions,
	}

	t.Execute(&engineCfg, engineConfigContext)

	return &ContainerRuntimeOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: bp.CRuntimeOptionsFile,
	}, nil
}

func (provisioner *BaseProvisioner) CompatibleWithMachine() bool {
	return provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID
}

func (bp *BaseProvisioner) SetOsReleaseInfo(info *OsRelease) {
	bp.OsReleaseInfo = info
}

// GetOsReleaseInfo returns a struct referring to the /etc/os-release file
// inside the linux machine
func (bp *BaseProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return bp.OsReleaseInfo, nil
}

// GetDriver returns a reference for the driver that we're making use of
func (bp *BaseProvisioner) GetDriver() drivers.Driver {
	return bp.Driver
}

// TODO:
func (bp *BaseProvisioner) ServiceAction(string, serviceaction.ServiceAction) error
func (bp *BaseProvisioner) PackageAction(string, pkgaction.PackageAction) error
func (bp *BaseProvisioner) ConfigureAuth() error
func (bp *BaseProvisioner) Provision(auth.Options, cruntime.Options) error

// DONE:
// func (bp *BaseProvisioner) Copy(assets.CopyableFile) error
// func (bp *BaseProvisioner) CopyFrom(assets.CopyableFile) error
// func (bp *BaseProvisioner) RemoveFile(assets.CopyableFile) error
// func (bp *BaseProvisioner) ReadableFile(sourcePath string) (assets.ReadableFile, error)
// func (bp *BaseProvisioner) WaitCmd(startedCmd *runner.StartedCmd) (*runner.RunResult, error)
// func (bp *BaseProvisioner) StartCmd(cmd *exec.Cmd) (*runner.StartedCmd, error)
// func (bp *BaseProvisioner) GetProvisionerName() string
// func (bp *BaseProvisioner) RunCmd(cmd *exec.Cmd) (*runner.RunResult, error)
// func (bp *BaseProvisioner) CompatibleWithMachine() bool
// func (bp *BaseProvisioner) GetCRuntime() string R
// func (bp *BaseProvisioner) GetCRuntimeOptionsDir() string
// func (bp *BaseProvisioner) GenerateCRuntimeOptions(interface{}) (*ContainerRuntimeOptions, error)
// func (bp *BaseProvisioner) GetOsReleaseInfo() (*OsRelease, error)
// func (bp *BaseProvisioner) SetOsReleaseInfo(*OsRelease)
