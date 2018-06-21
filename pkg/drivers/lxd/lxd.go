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
	URL    string
	server lxd.ContainerServer
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
	fmt.Println("LXD: create ")
	c, err := getConnection()
	if err != nil {
		return err
	}
	req := api.ContainersPost{
		Name: d.MachineName,
		Source: api.ContainerSource{
			Type:  "image",
			Alias: "ubuntu",
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
	reqState := api.ContainerStatePut{
		Action:  "start",
		Timeout: -1,
	}

	op, err = c.UpdateContainerState(d.MachineName, reqState, "")
	if err != nil {
		return errors.Wrap(err, "Error updating container state")
	}

	err = op.Wait()
	if err != nil {
		return errors.Wrap(err, "Error waiting for container state to be updated")
	}
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
	c, err := getConnection()
	if err != nil {
		return state.None, err
	}
	lxdState, etag, err := c.GetContainerState(d.MachineName)
	if err != nil {
		return state.None, errors.Wrap(err, "Error getting container state")
	}
	fmt.Printf("LXD: container state ETag: %s\n", etag)
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

func (d *Driver) Kill() error {
	fmt.Println("LXD: kill ")
	c, err := getConnection()
	if err != nil {
		return err
	}
	reqState := api.ContainerStatePut{
		Action:  "kill",
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

func (d *Driver) Remove() error {
	fmt.Println("LXD: delete ")
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

func (d *Driver) Restart() error {
	fmt.Println("LXD: restarting ")
	return nil
}

func (d *Driver) Start() error {
	fmt.Println("LXD: starting ")
	return d.UpdateState("start")
}

func (d *Driver) Stop() error {
	fmt.Println("LXD: stopping ")
	return d.UpdateState("stop")
}

func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("LXD: driver does not support ssh commands")
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
