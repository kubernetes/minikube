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

package lxd

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	lxd "github.com/lxc/lxd/client"
	"github.com/pkg/errors"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
)

const driverName = "lxd"

// lxd Driver is for running minikube inside LXD
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	URL string
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// PreCreateCheck checks for correct privileges and dependencies
func (d *Driver) PreCreateCheck() error {
	fmt.Println("LXD: pre create check")
	if _, err := getConnection(); err != nil {
		return errors.Wrap(err, "Error connecting to LXD API. Is your LXD service running?")
	}
	return nil
}

func (d *Driver) Create() error {
	fmt.Println("LXD: create ")
	// creation for the none driver is handled by commands.go
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetIP() (string, error) {
	fmt.Println("LXD: get ip ")
	return "127.0.0.1", nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	fmt.Println("LXD: get ssh hostname ")
	return "minikube", nil
}

func (d *Driver) GetSSHPort() (int, error) {
	fmt.Println("LXD: get ssh port ")
	return 2222, nil
}

func (d *Driver) GetURL() (string, error) {
	fmt.Println("LXD: get url ")
	return "lxd://minikube", nil
}

func (d *Driver) GetState() (state.State, error) {
	fmt.Println("LXD: get state ")
	return state.None, nil
}

func (d *Driver) Kill() error {
	fmt.Println("LXD: kill ")
	return nil
}

func (d *Driver) Remove() error {
	fmt.Println("LXD: delete ")
	return nil
}

func (d *Driver) Restart() error {
	fmt.Println("LXD: restarting ")
	return nil
}

func (d *Driver) Start() error {
	fmt.Println("LXD: starting ")
	return nil
}

func (d *Driver) Stop() error {
	fmt.Println("LXD: stopping ")
	return nil
}

func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("LXD: driver does not support ssh commands")
}

func getConnection() (lxd.ContainerServer, error) {
	c, err := lxd.ConnectLXDUnix("", nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}
