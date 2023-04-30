package errdriver

import (
	"fmt"
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
)

type Driver struct {
	Name string
}

type NotLoadable struct {
	Name string
}

func (e NotLoadable) Error() string {
	return fmt.Sprintf("Driver %q not found. Do you have the plugin binary accessible in your PATH?", e.Name)
}

func NewDriver(Name string) drivers.Driver {
	return &Driver{
		Name: Name,
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "not-found"
}

func (d *Driver) PreCreateCheck() error {
	return NotLoadable{d.Name}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return NotLoadable{d.Name}
}

func (d *Driver) GetURL() (string, error) {
	return "", NotLoadable{d.Name}
}

func (d *Driver) GetMachineName() string {
	return d.Name
}

func (d *Driver) GetIP() (string, error) {
	return "1.2.3.4", NotLoadable{d.Name}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "", NotLoadable{d.Name}
}

func (d *Driver) GetSSHKeyPath() string {
	return ""
}

func (d *Driver) GetSSHPort() (int, error) {
	return 0, NotLoadable{d.Name}
}

func (d *Driver) GetSSHUsername() string {
	return ""
}

func (d *Driver) GetMachineState() (state.State, error) {
	return state.Error, NotLoadable{d.Name}
}

func (d *Driver) CreateMachine() error {
	return NotLoadable{d.Name}
}

func (d *Driver) RemoveMachine() error {
	return NotLoadable{d.Name}
}

func (d *Driver) StartMachine() error {
	return NotLoadable{d.Name}
}

func (d *Driver) StopMachine() error {
	return NotLoadable{d.Name}
}

func (d *Driver) RestartMachine() error {
	return NotLoadable{d.Name}
}

func (d *Driver) KillMachine() error {
	return NotLoadable{d.Name}
}

func (d *Driver) RunCmd(*exec.Cmd) (*runner.RunResult, error) {
	return nil, NotLoadable{d.Name}
}

func (d *Driver) GetRunner() (runner.Runner, error) {
	return nil, NotLoadable{d.Name}
}
