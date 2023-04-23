package drivers

import (
	"errors"
	"path/filepath"

	"k8s.io/minikube/pkg/libmachine/runner"
)

const (
	DefaultSSHUser          = "root"
	DefaultSSHPort          = 22
	DefaultEngineInstallURL = "https://get.docker.com"
)

// BaseDriver - Embed this struct into drivers to provide the common set
// of fields and functions.
type BaseDriver struct {
	IPAddress   string
	MachineName string
	SSHUser     string
	SSHPort     int
	SSHKeyPath  string
	StorePath   string
	Runner      runner.Runner
}

func (d *BaseDriver) RunCommand(args string) (string, error) {
	return d.Runner.RunCommand(args)
}

// DriverName returns the name of the driver
func (d *BaseDriver) DriverName() string {
	return "unknown"
}

// GetMachineName returns the machine name
func (d *BaseDriver) GetMachineName() string {
	return d.MachineName
}

// GetIP returns the ip
func (d *BaseDriver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", errors.New("IP address is not set")
	}
	return d.IPAddress, nil
}

// PreCreateCheck is called to enforce pre-creation steps
func (d *BaseDriver) PreCreateCheck() error {
	return nil
}

// ResolveStorePath returns the store path where the machine is
func (d *BaseDriver) ResolveStorePath(file string) string {
	return filepath.Join(d.StorePath, "machines", d.MachineName, file)
}

func EngineInstallURLFlagSet(flags DriverOptions) bool {
	return EngineInstallURLSet(flags.String("engine-install-url"))
}

func EngineInstallURLSet(url string) bool {
	return url != DefaultEngineInstallURL && url != ""
}
