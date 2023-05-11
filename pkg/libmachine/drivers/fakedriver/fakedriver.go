/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

// x7TODO: I can easily see this inside the libmachine/drivers folder...
package fakedriver

import (
	"fmt"
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	MockState state.State
	MockIP    string
	MockName  string
	Runner    *runner.FakeCommandRunner
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "Driver"
}

func (d *Driver) SetConfigFromFlags(_ drivers.DriverOptions) error {
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
		return "", fmt.Errorf("unable to get ip")
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

func (d *Driver) GetMachineState() (state.State, error) {
	return d.MockState, nil
}

func (d *Driver) CreateMachine() error {
	return nil
}

func (d *Driver) StartMachine() error {
	d.MockState = state.Running
	return nil
}

func (d *Driver) StopMachine() error {
	d.MockState = state.Stopped
	return nil
}

func (d *Driver) RestartMachine() error {
	d.MockState = state.Running
	return nil
}

func (d *Driver) KillMachine() error {
	d.MockState = state.Stopped
	return nil
}

func (d *Driver) RemoveMachine() error {
	return nil
}

func (d *Driver) Upgrade() error {
	return nil
}

func (d *Driver) RunCmd(cmd *exec.Cmd) (*runner.RunResult, error) {
	if d.Runner == nil {
		_, _ = d.GetRunner()
	}
	return d.Runner.RunCmd(cmd)
}

func (d *Driver) GetRunner() (runner.Runner, error) {
	if d.Runner == nil {
		d.Runner = runner.NewFakeCommandRunner()
	}

	return d.Runner, nil
}

func (d *Driver) SetCmdOutput(cmdToOut map[string]string) {
	if d.Runner == nil {
		_, _ = d.GetRunner()
	}
	d.Runner.SetCommandToOutput(cmdToOut)
}

func (d *Driver) IsISOBased() bool {
	return false
}

func (d *Driver) IsContainerBased() bool {
	return false
}

func (d *Driver) IsManaged() bool {
	return true
}
