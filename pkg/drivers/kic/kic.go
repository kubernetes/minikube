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
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/net"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

// Driver is a driver designed to run kubeadm w/o VM management, and assumes systemctl.
// https://minikube.sigs.k8s.io/docs/reference/drivers/none/
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	URL     string
	runtime cruntime.Manager
	exec    command.Runner
}

// Config is configuration for the kic driver
type Config struct {
	MachineName      string
	StorePath        string
	ContainerRuntime string
}

// NewDriver returns a fully configured None driver
func NewDriver(c Config) *Driver {
	runner := &command.ExecRunner{}
	runtime, err := cruntime.New(cruntime.Config{Type: c.ContainerRuntime, Runner: runner})
	// Libraries shouldn't panic, but there is no way for drivers to return error :(
	if err != nil {
		glog.Fatalf("unable to create container runtime: %v", err)
	}
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: c.MachineName,
			StorePath:   c.StorePath,
		},
		runtime: runtime,
		exec:    runner,
	}
}

// PreCreateCheck checks for correct privileges and dependencies
func (d *Driver) PreCreateCheck() error {
	return fmt.Errorf("driver does not support ssh commands")
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	// creation for the none driver is handled by commands.go
	return fmt.Errorf("driver does not support ssh commands")
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "kic"
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
	return "", fmt.Errorf("driver does not support ssh commands")
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	return 0, fmt.Errorf("driver does not support ssh commands")
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	return "tcp://1.2.3.4:2376", fmt.Errorf("not implemented for kic yet")
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	return state.Stopped, fmt.Errorf("not implemented for kic yet")
}

// Kill stops a host forcefully, including any containers that we are managing.
func (d *Driver) Kill() error {
	return fmt.Errorf("not implemented for kic yet")
}

func (d *Driver) Remove() error {
	return fmt.Errorf("not implemented for kic yet")
}

// Restart a host
func (d *Driver) Restart() error {
	return fmt.Errorf("not implemented for kic yet")
}

// Start a host
func (d *Driver) Start() error {
	return fmt.Errorf("not implemented for kic yet")
}

// Stop a host gracefully, including any containers that we are managing.
func (d *Driver) Stop() error {
	return fmt.Errorf("not implemented for kic yet")
}

// RunSSHCommandFromDriver implements direct ssh control to the driver
func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("not implemented for kic yet")
}
