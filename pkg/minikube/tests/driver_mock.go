/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tests

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// MockDriver is a struct used to mock out libmachine.Driver
type MockDriver struct {
	drivers.BaseDriver
	CurrentState state.State
	RemoveError  bool
	HostError    bool
	Port         int
	IP           string
	T            *testing.T
}

// Logf logs mock interactions
func (d *MockDriver) Logf(format string, args ...interface{}) {
	if d.T == nil {
		glog.Infof(format, args...)
		return
	}
	d.T.Logf(format, args...)
}

// Create creates a MockDriver instance
func (d *MockDriver) Create() error {
	d.Logf("MockDriver.Create")
	d.CurrentState = state.Running
	return nil
}

// GetIP returns the IP address
func (d *MockDriver) GetIP() (string, error) {
	d.Logf("MockDriver.GetIP")
	if d.IP != "" {
		return d.IP, nil
	}
	if d.BaseDriver.IPAddress != "" {
		return d.BaseDriver.IPAddress, nil
	}
	return "127.0.0.1", nil
}

// GetCreateFlags returns the flags used to create a MockDriver
func (d *MockDriver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{}
}

// GetSSHPort returns the SSH port
func (d *MockDriver) GetSSHPort() (int, error) {
	return d.Port, nil
}

// GetSSHHostname returns the hostname for SSH
func (d *MockDriver) GetSSHHostname() (string, error) {
	if d.HostError {
		return "", errors.New("error getting host")
	}
	return "localhost", nil
}

// GetSSHKeyPath returns the key path for SSH
func (d *MockDriver) GetSSHKeyPath() string {
	return d.BaseDriver.SSHKeyPath
}

// GetState returns the state of the driver
func (d *MockDriver) GetState() (state.State, error) {
	d.Logf("MockDriver.GetState: %v", d.CurrentState)
	return d.CurrentState, nil
}

// GetURL returns the URL of the driver
func (d *MockDriver) GetURL() (string, error) {
	return "", nil
}

// Kill kills the machine
func (d *MockDriver) Kill() error {
	d.Logf("MockDriver.Kill")
	d.CurrentState = state.Stopped
	return nil
}

// Remove removes the machine
func (d *MockDriver) Remove() error {
	d.Logf("MockDriver.Remove")
	if d.RemoveError {
		return errors.New("error deleting machine")
	}
	return nil
}

// Restart restarts the machine
func (d *MockDriver) Restart() error {
	d.Logf("MockDriver.Restart")
	d.CurrentState = state.Running
	return nil
}

// SetConfigFromFlags sets the machine config
func (d *MockDriver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	return nil
}

// Start starts the machine
func (d *MockDriver) Start() error {
	d.Logf("MockDriver.Start")
	d.CurrentState = state.Running
	return nil
}

// Stop stops the machine
func (d *MockDriver) Stop() error {
	d.Logf("MockDriver.Stop")
	d.CurrentState = state.Stopped
	return nil
}

// DriverName returns the name of the driver
func (d *MockDriver) DriverName() string {
	d.Logf("MockDriver.Name")
	return "mock"
}
