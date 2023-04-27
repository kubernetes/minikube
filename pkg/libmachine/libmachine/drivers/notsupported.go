package drivers

import (
	"fmt"
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/assets"
)

type DriverNotSupported struct {
	*BaseDriver
	Name string
}

type NotSupported struct {
	DriverName string
}

func (e NotSupported) Error() string {
	return fmt.Sprintf("Driver %q not supported on this platform.", e.DriverName)
}

// NewDriverNotSupported creates a placeholder Driver that replaces
// a driver that is not supported on a given platform. eg fusion on linux.
func NewDriverNotSupported(driverName, hostName, storePath string) Driver {
	return &DriverNotSupported{
		BaseDriver: &BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
		Name: driverName,
	}
}

func (d *DriverNotSupported) DriverName() string {
	return d.Name
}

func (d *DriverNotSupported) PreCreateCheck() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) GetCreateFlags() []mcnflag.Flag {
	return nil
}

func (d *DriverNotSupported) SetConfigFromFlags(flags DriverOptions) error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) GetURL() (string, error) {
	return "", NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) GetSSHHostname() (string, error) {
	return "", NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) GetMachineState() (state.State, error) {
	return state.Error, NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) CreateMachine() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) RemoveMachine() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) StartMachine() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) StopMachine() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) RestartMachine() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) KillMachine() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) RunCmd(*exec.Cmd) (*runner.RunResult, error) {
	return nil, NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) StartCmd(*exec.Cmd) (*runner.StartedCmd, error) {
	return nil, NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) WaitCmd(*runner.StartedCmd) (*runner.RunResult, error) {
	return nil, NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) CopyFile(assets.CopyableFile) error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) CopyFileFrom(assets.CopyableFile) error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) RemoveFile(assets.CopyableFile) error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) ReadableFile(string) (assets.ReadableFile, error) {
	return nil, NotSupported{d.DriverName()}
}
