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

func (d *Driver) IsContainerBased() bool {
	return false
}

func (d *Driver) IsISOBased() bool {
	return false
}

func (d *Driver) IsManaged() bool {
	return true
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
