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
	"log"
	"os/user"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"github.com/pkg/errors"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
)

const (
	driverName            = "lxd"
	connectionErrStr      = "Error connecting to LXD API. Is your LXD service running?"
	createContainerErrStr = "Error creating LXD container."
)

// lxd Driver is for running minikube inside LXD
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
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
	if _, err := getConnection(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Create() error {
	c, err := getConnection()
	if err != nil {
		return err
	}
	req := api.ContainersPost{
		Name: d.MachineName,
		Source: api.ContainerSource{
			Type:  "image",
			Alias: "base-minikube",
		},
	}
	op, err := c.CreateContainer(req)
	if err != nil {
		return errors.Wrap(err, "Error creating container")
	}
	err = op.Wait()
	if err != nil {
		return errors.Wrap(err, "Error waiting for container to be created")
	}
	return d.Start()
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetIP() (string, error) {
	c, err := getConnection()
	if err != nil {
		return "", err
	}
	lxdState, _, err := c.GetContainerState(d.MachineName)
	if err != nil {
		return "", errors.Wrap(err, "Error getting container state")
	}

	eth0, ok := lxdState.Network["eth0"]
	if !ok {
		return "", errors.Wrap(err, "eth0 network doesn't exist")
	}
	for _, address := range eth0.Addresses {
		if address.Family == "inet" {
			return address.Address, nil
		}
	}
	return "", errors.New("couldn't find IPV4 address for container")
}

func (d *Driver) GetSSHKeyPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s/.ssh/id_rsa_lxd", usr.HomeDir)
}

func (d *Driver) GetSSHUsername() string {
	return "ubuntu"
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", errors.Wrap(err, "getting URL, could not get IP")
	}
	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetState() (state.State, error) {
	c, err := getConnection()
	if err != nil {
		return state.None, err
	}
	lxdState, _, err := c.GetContainerState(d.MachineName)
	if err != nil {
		return state.None, errors.Wrap(err, "Error getting container state")
	}

	var mkState state.State
	switch lxdState.StatusCode {
	case api.Running:
		mkState = state.Running
	case api.Freezing, api.Frozen:
		mkState = state.Paused
	case api.Stopped:
		mkState = state.Stopped
	case api.Stopping, api.Aborting:
		mkState = state.Stopping
	case api.OperationCreated, api.Started, api.Pending, api.Starting, api.Thawed:
		mkState = state.Starting
	case api.Error, api.Failure:
		mkState = state.Error
	default:
		mkState = state.None
	}
	return mkState, nil
}

func (d *Driver) Restart() error {
	return d.UpdateState("restart")
}

func (d *Driver) Start() error {
	return d.UpdateState("start")
}

func (d *Driver) Stop() error {
	return d.UpdateState("stop")
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Remove() error {
	d.Stop()
	c, err := getConnection()
	if err != nil {
		return err
	}

	op, err := c.DeleteContainer(d.MachineName)
	if err != nil {
		return errors.Wrap(err, "Error deleting container")
	}

	err = op.Wait()
	if err != nil {
		return errors.Wrap(err, "Error waiting for container to be deleted")
	}
	return nil
}

func getConnection() (lxd.ContainerServer, error) {
	c, err := lxd.ConnectLXDUnix("", nil)
	if err != nil {
		return nil, errors.Wrap(err, connectionErrStr)
	}
	return c, nil
}

func (d *Driver) UpdateState(state string) error {
	c, err := getConnection()
	if err != nil {
		return err
	}
	reqState := api.ContainerStatePut{
		Action:  state,
		Timeout: -1,
	}

	op, err := c.UpdateContainerState(d.MachineName, reqState, "")
	if err != nil {
		return errors.Wrap(err, "Error updating container state")
	}

	err = op.Wait()
	if err != nil {
		return errors.Wrap(err, "Error waiting for container state to be updated")
	}
	return nil
}
