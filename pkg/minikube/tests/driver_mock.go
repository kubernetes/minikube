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

	"k8s.io/minikube/pkg/minikube/constants"

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
func (driver *MockDriver) Logf(format string, args ...interface{}) {
	if driver.T == nil {
		glog.Infof(format, args...)
		return
	}
	driver.T.Logf(format, args...)
}

// Create creates a MockDriver instance
func (driver *MockDriver) Create() error {
	driver.Logf("MockDriver.Create")
	driver.CurrentState = state.Running
	return nil
}

// GetIP returns the IP address
func (driver *MockDriver) GetIP() (string, error) {
	driver.Logf("MockDriver.GetIP")
	if driver.IP != "" {
		return driver.IP, nil
	}
	if driver.BaseDriver.IPAddress != "" {
		return driver.BaseDriver.IPAddress, nil
	}
	return "127.0.0.1", nil
}

// GetCreateFlags returns the flags used to create a MockDriver
func (driver *MockDriver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{}
}

// GetSSHPort returns the SSH port
func (driver *MockDriver) GetSSHPort() (int, error) {
	return driver.Port, nil
}

// GetSSHHostname returns the hostname for SSH
func (driver *MockDriver) GetSSHHostname() (string, error) {
	if driver.HostError {
		return "", errors.New("error getting host")
	}
	return "localhost", nil
}

// GetSSHKeyPath returns the key path for SSH
func (driver *MockDriver) GetSSHKeyPath() string {
	return driver.BaseDriver.SSHKeyPath
}

// GetState returns the state of the driver
func (driver *MockDriver) GetState() (state.State, error) {
	driver.Logf("MockDriver.GetState: %v", driver.CurrentState)
	return driver.CurrentState, nil
}

// GetURL returns the URL of the driver
func (driver *MockDriver) GetURL() (string, error) {
	return "", nil
}

// Kill kills the machine
func (driver *MockDriver) Kill() error {
	driver.Logf("MockDriver.Kill")
	driver.CurrentState = state.Stopped
	return nil
}

// Remove removes the machine
func (driver *MockDriver) Remove() error {
	driver.Logf("MockDriver.Remove")
	if driver.RemoveError {
		return errors.New("error deleting machine")
	}
	return nil
}

// Restart restarts the machine
func (driver *MockDriver) Restart() error {
	driver.Logf("MockDriver.Restart")
	driver.CurrentState = state.Running
	return nil
}

// SetConfigFromFlags sets the machine config
func (driver *MockDriver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	return nil
}

// Start starts the machine
func (driver *MockDriver) Start() error {
	driver.Logf("MockDriver.Start")
	driver.CurrentState = state.Running
	return nil
}

// Stop stops the machine
func (driver *MockDriver) Stop() error {
	driver.Logf("MockDriver.Stop")
	driver.CurrentState = state.Stopped
	return nil
}

// DriverName returns the name of the driver
func (driver *MockDriver) DriverName() string {
	driver.Logf("MockDriver.Name")
	return constants.DriverMock
}
