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
	"runtime"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"

	"k8s.io/klog/v2"
)

// MockDriver is a struct used to mock out libmachine.Driver
type MockDriver struct {
	drivers.BaseDriver
	CurrentState  state.State
	RemoveError   bool
	NotExistError bool
	HostError     bool
	Port          int
	IP            string
	T             *testing.T
}

// Logf logs mock interactions
func (d *MockDriver) Logf(format string, args ...interface{}) {
	if d.T == nil {
		klog.Infof(format, args...)
		return
	}
	d.T.Logf(format, args...)
}

// Create creates a MockDriver instance
func (d *MockDriver) Create() error {
	d.Logf("MockDriver.Create")
	if d.NotExistError {
		d.Logf("MockDriver.Create but machine does not exist")
		d.CurrentState = state.Error
		return nil
	}
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
	_, file, no, _ := runtime.Caller(2)
	d.Logf("MockDriver.GetState called from %s#%d: returning %q", file, no, d.CurrentState)

	// NOTE: this logic is questionable
	if d.NotExistError && d.CurrentState != state.Stopped && d.CurrentState != state.None {
		d.CurrentState = state.Error
		d.Logf("mock NotExistError set, setting state=%s err=%v", d.CurrentState, constants.ErrMachineMissing)
		return d.CurrentState, constants.ErrMachineMissing
	}
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
	d.NotExistError = false
	return nil
}

// Restart restarts the machine
func (d *MockDriver) Restart() error {
	d.Logf("MockDriver.Restart, setting CurrentState=%s", state.Running)
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
