package tests

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

// MockAPI is a struct used to mock out libmachine.API
type MockAPI struct {
	Hosts []*host.Host
}

// MockDriver is a struct used to mock out libmachine.Driver
type MockDriver struct {
	drivers.BaseDriver
	CurrentState state.State
}

// Create creates a MockDriver instance
func (driver *MockDriver) Create() error {
	driver.CurrentState = state.Running
	return nil
}

// GetCreateFlags returns the flags used to create a MockDriver
func (driver *MockDriver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{}
}

// GetSSHHostname returns the hostname for SSH
func (driver *MockDriver) GetSSHHostname() (string, error) {
	return "", nil
}

// GetState returns the state of the driver
func (driver *MockDriver) GetState() (state.State, error) {
	return driver.CurrentState, nil
}

// GetURL returns the URL of the driver
func (driver *MockDriver) GetURL() (string, error) {
	return "", nil
}

// Kill kills the machine
func (driver *MockDriver) Kill() error {
	driver.CurrentState = state.Stopped
	return nil
}

// Remove removes the machine
func (driver *MockDriver) Remove() error {
	return nil
}

// Restart restarts the machine
func (driver *MockDriver) Restart() error {
	driver.CurrentState = state.Running
	return nil
}

// SetConfigFromFlags sets the machine config
func (driver *MockDriver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	return nil
}

// Start starts the machine
func (driver *MockDriver) Start() error {
	driver.CurrentState = state.Running
	return nil
}

// Stop stops the machine
func (driver *MockDriver) Stop() error {
	driver.CurrentState = state.Stopped
	return nil
}
