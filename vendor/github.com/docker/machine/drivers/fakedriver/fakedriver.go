package fakedriver

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	MockState state.State
	MockIP    string
	MockName  string
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "Driver"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetMachineName() string {
	return d.MockName
}

func (d *Driver) GetIP() (string, error) {
	if d.MockState == state.Error {
		return "", fmt.Errorf("Unable to get ip")
	}
	if d.MockState == state.Timeout {
		select {} // Loop forever
	}
	if d.MockState != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}
	return d.MockIP, nil
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

func (d *Driver) GetState() (state.State, error) {
	return d.MockState, nil
}

func (d *Driver) Create() error {
	return nil
}

func (d *Driver) Start() error {
	d.MockState = state.Running
	return nil
}

func (d *Driver) Stop() error {
	d.MockState = state.Stopped
	return nil
}

func (d *Driver) Restart() error {
	d.MockState = state.Running
	return nil
}

func (d *Driver) Kill() error {
	d.MockState = state.Stopped
	return nil
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Upgrade() error {
	return nil
}
