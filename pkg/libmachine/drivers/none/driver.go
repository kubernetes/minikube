package none

import (
	"fmt"
	neturl "net/url"
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
)

const driverName = "none"

// Driver is the driver used when no driver is selected. It is used to
// connect to existing Docker hosts by specifying the URL of the host as
// an option.
type Driver struct {
	*drivers.BaseDriver
	URL string
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:  "url",
			Usage: "URL of host when no driver is selected",
			Value: "",
		},
	}
}

func (d *Driver) CreateMachine() error {
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetIP() (string, error) {
	return d.IPAddress, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "", nil
}

func (d *Driver) GetSSHKeyPath() string {
	return ""
}

func (d *Driver) GetSSHPort() (int, error) {
	return 0, nil
}

func (d *Driver) GetSSHUsername() string {
	return ""
}

func (d *Driver) GetURL() (string, error) {
	return d.URL, nil
}

func (d *Driver) GetMachineState() (state.State, error) {
	return state.Running, nil
}

func (d *Driver) KillMachine() error {
	return fmt.Errorf("hosts without a driver cannot be killed")
}

func (d *Driver) RemoveMachine() error {
	return nil
}

func (d *Driver) RestartMachine() error {
	return fmt.Errorf("hosts without a driver cannot be restarted")
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	url := flags.String("url")

	if url == "" {
		return fmt.Errorf("--url option is required when no driver is selected")
	}

	d.URL = url
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}

	d.IPAddress = u.Host
	return nil
}

func (d *Driver) StartMachine() error {
	return fmt.Errorf("hosts without a driver cannot be started")
}

func (d *Driver) StopMachine() error {
	return fmt.Errorf("hosts without a driver cannot be stopped")
}

func (d *Driver) RunCmd(*exec.Cmd) (*runner.RunResult, error) {
	return nil, fmt.Errorf("cannot run a commnd without a driver")
}

func (d *Driver) GetRunner() (runner.Runner, error) {
	return nil, fmt.Errorf("cannot get a runner inside the machine without a driver")
}
