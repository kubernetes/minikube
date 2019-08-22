/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package kic

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/net"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

const driverName = constants.DriverKic

// Driver is a driver designed to run kubeadm inside a container
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	URL       string
	runtime   cruntime.Manager
	exec      command.Runner
	OciClient string // docker or podman
	KicImage  string
}

// Config is configuration for the kic driver
type Config struct {
	MachineName      string
	StorePath        string
	ContainerRuntime string
	OciClient        string
	KicImage         string // Image used for nodes
}

// NewDriver returns a fully configured None driver
func NewDriver(c Config) *Driver {
	runner := &command.OciRunner{} // MEDYA:TODO pass the node selector
	// runtime, err := cruntime.New(cruntime.Config{Type: c.ContainerRuntime, Runner: runner})
	// Libraries shouldn't panic, but there is no way for drivers to return error :(
	// if err != nil {
	// 	glog.Fatalf("unable to create container runtime: %v", err)
	// }
	d := &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: c.MachineName,
			StorePath:   c.StorePath,
		},
		CommonDriver: &pkgdrivers.CommonDriver{},
		exec:         runner,
		KicImage:     c.KicImage,
		OciClient:    c.OciClient,
	}

	fmt.Printf("MEDYA: Creating a New Driver %v", d)
	return d
}

// PreCreateCheck checks for correct privileges and dependencies
func (d *Driver) PreCreateCheck() error {
	err := d.runtime.Available()
	if err != nil {
		return errors.Wrap(err, "pre-check for kic driver")
	}
	// TODO check if oci is in path
	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	// creation for the none driver is handled by commands.go
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

// GetIP returns an IP or hostname that this host is available at
func (d *Driver) GetIP() (string, error) {
	ip, err := net.ChooseBindAddress(nil)
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	// TODO: change this it does support
	return "", fmt.Errorf("driver does not support ssh commands")
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	// TODO: change this it does support
	return 0, fmt.Errorf("driver does not support ssh commands")
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	// TODO: add later
	return state.Running, nil
}

// Kill stops a host forcefully, including any containers that we are managing.
func (d *Driver) Kill() error {
	// TODO later
	return nil
}

// Remove a host, including any data which may have been written by it.
func (d *Driver) Remove() error {
	// TODO: remove the image from cache maybe ?
	return nil
}

// Restart a host
func (d *Driver) Restart() error {
	// TODO: later
	return nil
}

// Start a host
func (d *Driver) Start() error {
	var err error
	d.IPAddress, err = d.GetIP()
	if err != nil {
		return err
	}
	d.URL, err = d.GetURL()
	if err != nil {
		return err
	}
	return nil
}

// Stop a host gracefully, including any containers that we are managing.
func (d *Driver) Stop() error {
	// pause
	return nil
}
