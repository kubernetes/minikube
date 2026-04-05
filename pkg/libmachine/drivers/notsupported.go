/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package drivers

import (
	"fmt"

	"k8s.io/minikube/pkg/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/state"
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

func (d *DriverNotSupported) GetState() (state.State, error) {
	return state.Error, NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Create() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Remove() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Start() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Stop() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Restart() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Kill() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Upgrade() error {
	return NotSupported{d.DriverName()}
}
