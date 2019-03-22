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

package none

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/net"
	pkgdrivers "k8s.io/minikube/pkg/drivers" // TODO(tstromberg): Extract CommandRunner into its own package
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

const driverName = "none"

// cleanupPaths are paths to be removed by cleanup, and are used by both kubeadm and minikube.
var cleanupPaths = []string{
	"/data/minikube",
	"/etc/kubernetes/manifests",
	"/var/lib/minikube",
}

// Driver is a driver designed to run kubeadm w/o VM management, and assumes systemctl.
// https://github.com/kubernetes/minikube/blob/master/docs/vmdriver-none.md
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	URL     string
	runtime cruntime.Manager
	exec    bootstrapper.CommandRunner
}

// Config is configuration for the None driver
type Config struct {
	MachineName      string
	StorePath        string
	ContainerRuntime string
}

// NewDriver returns a fully configured None driver
func NewDriver(c Config) *Driver {
	runner := &bootstrapper.ExecRunner{}
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
	return d.runtime.Available()
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
	return "", fmt.Errorf("driver does not support ssh commands")
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
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
	if err := checkKubelet(d.exec); err != nil {
		glog.Infof("kubelet not running: %v", err)
		return state.Stopped, nil
	}
	return state.Running, nil
}

// Kill stops a host forcefully, including any containers that we are managing.
func (d *Driver) Kill() error {
	if err := stopKubelet(d.exec); err != nil {
		return errors.Wrap(err, "kubelet")
	}

	// First try to gracefully stop containers
	containers, err := d.runtime.ListContainers("")
	if err != nil {
		return errors.Wrap(err, "containers")
	}
	if len(containers) == 0 {
		return nil
	}
	// Try to be graceful before sending SIGKILL everywhere.
	if err := d.runtime.StopContainers(containers); err != nil {
		return errors.Wrap(err, "stop")
	}

	containers, err = d.runtime.ListContainers("")
	if err != nil {
		return errors.Wrap(err, "containers")
	}
	if len(containers) == 0 {
		return nil
	}
	if err := d.runtime.KillContainers(containers); err != nil {
		return errors.Wrap(err, "kill")
	}
	return nil
}

// Remove a host, including any data which may have been written by it.
func (d *Driver) Remove() error {
	if err := d.Kill(); err != nil {
		return errors.Wrap(err, "kill")
	}
	glog.Infof("Removing: %s", cleanupPaths)
	cmd := fmt.Sprintf("sudo rm -rf %s", strings.Join(cleanupPaths, " "))
	if err := d.exec.Run(cmd); err != nil {
		glog.Errorf("cleanup incomplete: %v", err)
	}
	return nil
}

// Restart a host
func (d *Driver) Restart() error {
	return restartKubelet(d.exec)
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
	if err := stopKubelet(d.exec); err != nil {
		return err
	}
	containers, err := d.runtime.ListContainers("")
	if err != nil {
		return errors.Wrap(err, "containers")
	}
	if len(containers) > 0 {
		if err := d.runtime.StopContainers(containers); err != nil {
			return errors.Wrap(err, "stop")
		}
	}
	return nil
}

// RunSSHCommandFromDriver implements direct ssh control to the driver
func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("driver does not support ssh commands")
}

// stopKubelet idempotently stops the kubelet
func stopKubelet(exec bootstrapper.CommandRunner) error {
	glog.Infof("stopping kubelet.service ...")
	return exec.Run("sudo systemctl stop kubelet.service")
}

// restartKubelet restarts the kubelet
func restartKubelet(exec bootstrapper.CommandRunner) error {
	glog.Infof("restarting kubelet.service ...")
	return exec.Run("sudo systemctl restart kubelet.service")
}

// checkKubelet returns an error if the kubelet is not running.
func checkKubelet(exec bootstrapper.CommandRunner) error {
	glog.Infof("checking for running kubelet ...")
	return exec.Run("systemctl is-active --quiet service kubelet")
}
