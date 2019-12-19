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
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/drivers/kic/node"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/command"
)

// https://minikube.sigs.k8s.io/docs/reference/drivers/kic/
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	URL        string
	exec       command.Runner
	NodeConfig Config
	OCIBinary  string // docker,podman
}

// Config is configuration for the kic driver used by registry
type Config struct {
	MachineName   string            // maps to the container name being created
	CPU           int               // Number of CPU cores assigned to the container
	Memory        int               // max memory in MB
	StorePath     string            // lib machine store path
	OCIBinary     string            // oci tool to use (docker, podman,...)
	ImageDigest   string            // image name with sha to use for the node
	APIServerPort int32             // port to connect to forward from container to user's machine
	Mounts        []oci.Mount       // mounts
	PortMappings  []oci.PortMapping // container port mappings
	Envs          map[string]string // key,value of environment variables passed to the node
}

// NewDriver returns a fully configured Kic driver
func NewDriver(c Config) *Driver {
	d := &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: c.MachineName,
			StorePath:   c.StorePath,
		},
		exec:       command.NewKICRunner(c.MachineName, c.OCIBinary),
		NodeConfig: c,
	}
	return d
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	params := node.CreateConfig{
		Name:         d.NodeConfig.MachineName,
		Image:        d.NodeConfig.ImageDigest,
		ClusterLabel: node.ClusterLabelKey + "=" + d.MachineName,
		CPUs:         strconv.Itoa(d.NodeConfig.CPU),
		Memory:       strconv.Itoa(d.NodeConfig.Memory) + "mb",
		Envs:         d.NodeConfig.Envs,
		ExtraArgs:    []string{"--expose", fmt.Sprintf("%d", d.NodeConfig.APIServerPort)},
		OCIBinary:    d.NodeConfig.OCIBinary,
	}

	// control plane specific options
	params.PortMappings = append(params.PortMappings, oci.PortMapping{
		ListenAddress: "127.0.0.1",
		HostPort:      d.NodeConfig.APIServerPort,
		ContainerPort: 6443,
	})

	_, err := node.CreateNode(params)
	if err != nil {
		return errors.Wrap(err, "create kic node")
	}
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	if d.NodeConfig.OCIBinary == "podman" {
		return "podman"
	}
	return "docker"
}

// GetIP returns an IP or hostname that this host is available at
func (d *Driver) GetIP() (string, error) {
	node, err := node.Find(d.OCIBinary, d.MachineName, d.exec)
	if err != nil {
		return "", fmt.Errorf("ip not found for nil node")
	}
	ip, _, err := node.IP()
	return ip, err
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return "", fmt.Errorf("driver does not have SSHHostName")
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	return 0, fmt.Errorf("driver does not support GetSSHPort")
}

// GetURL returns ip of the container running kic control-panel
func (d *Driver) GetURL() (string, error) {
	return d.GetIP()
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	cmd := exec.Command(d.NodeConfig.OCIBinary, "inspect", "-f", "{{.State.Status}}", d.MachineName)
	out, err := cmd.CombinedOutput()
	o := strings.Trim(string(out), "\n")
	if err != nil {
		return state.Error, errors.Wrapf(err, "error stop node %s", d.MachineName)
	}
	if o == "running" {
		return state.Running, nil
	}
	if o == "exited" {
		return state.Stopped, nil
	}
	if o == "paused" {
		return state.Paused, nil
	}
	if o == "restarting" {
		return state.Starting, nil
	}
	if o == "dead" {
		return state.Error, nil
	}
	return state.None, fmt.Errorf("unknown state")

}

// Kill stops a host forcefully, including any containers that we are managing.
func (d *Driver) Kill() error {
	cmd := exec.Command(d.NodeConfig.OCIBinary, "kill", d.MachineName)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "killing kic node %s", d.MachineName)
	}
	return nil
}

// Remove will delete the Kic Node Container
func (d *Driver) Remove() error {
	if _, err := d.nodeID(d.MachineName); err != nil {
		return errors.Wrapf(err, "not found node %s", d.MachineName)
	}
	cmd := exec.Command(d.NodeConfig.OCIBinary, "rm", "-f", "-v", d.MachineName)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "error removing node %s", d.MachineName)
	}
	return nil
}

// Restart a host
func (d *Driver) Restart() error {
	s, err := d.GetState()
	if err != nil {
		return errors.Wrap(err, "get kic state")
	}
	switch s {
	case state.Paused:
		return d.Unpause()
	case state.Stopped:
		return d.Start()
	case state.Running, state.Error:
		if err = d.Stop(); err != nil {
			return fmt.Errorf("restarting a kic stop phase %v", err)
		}
		if err = d.Start(); err != nil {
			return fmt.Errorf("restarting a kic start phase %v", err)
		}
		return nil
	}

	return fmt.Errorf("restarted not implemented for kic state %s yet", s)
}

// Unpause a kic container
func (d *Driver) Unpause() error {
	cmd := exec.Command(d.NodeConfig.OCIBinary, "pause", d.MachineName)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "unpausing %s", d.MachineName)
	}
	return nil
}

// Start a _stopped_ kic container
// not meant to be used for Create().
func (d *Driver) Start() error {
	s, err := d.GetState()
	if err != nil {
		return errors.Wrap(err, "get kic state")
	}
	if s == state.Stopped {
		cmd := exec.Command(d.NodeConfig.OCIBinary, "start", d.MachineName)
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "starting a stopped kic node %s", d.MachineName)
		}
		return nil
	}
	return fmt.Errorf("cant start a not-stopped (%s) kic node", s)
}

// Stop a host gracefully, including any containers that we are managing.
func (d *Driver) Stop() error {
	cmd := exec.Command(d.NodeConfig.OCIBinary, "stop", d.MachineName)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "stopping %s", d.MachineName)
	}
	return nil
}

// RunSSHCommandFromDriver implements direct ssh control to the driver
func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("driver does not support RunSSHCommandFromDriver commands")
}

// looks up for a container node by name, will return error if not found.
func (d *Driver) nodeID(nameOrID string) (string, error) {
	cmd := exec.Command(d.NodeConfig.OCIBinary, "inspect", "-f", "{{.Id}}", nameOrID)
	id, err := cmd.CombinedOutput()
	if err != nil {
		id = []byte{}
	}
	return string(id), err
}

func ImageForVersion(ver string) (string, error) {
	switch ver {
	case "v1.11.10":
		return "medyagh/kic:v1.11.10@sha256:23bb7f5e8dd2232ec829132172e87f7b9d8de65269630989e7dac1e0fe993b74", nil
	case "v1.12.8":
		return "medyagh/kic:v1.12.8@sha256:c74bc5f3efe3539f6e1ad7f11bf7c09f3091c0547cb28071f4e43067053e5898", nil
	case "v1.12.9":
		return "medyagh/kic:v1.12.9@sha256:ff82f58e18dcb22174e8eb09dae14f7edd82d91a83c7ef19e33298d0eba6a0e3", nil
	case "v1.12.10":
		return "medyagh/kic:v1.12.10@sha256:2d174bae7c20698e59791e7cca9b6db234053d1a92a009d5bb124e482540c70b", nil
	case "v1.13.6":
		return "medyagh/kic:v1.13.6@sha256:cf63e50f824fe17b90374d38d64c5964eb9fe6b3692669e1201fcf4b29af4964", nil
	case "v1.13.7":
		return "medyagh/kic:v1.13.7@sha256:1a6a5e1c7534cf3012655e99df680496df9bcf0791a304adb00617d5061233fa", nil
	case "v1.14.3":
		return "medyagh/kic:v1.14.3@sha256:cebec21f6af23d5dfa3465b88ddf4a1acb94c2c20a0a6ff8cc1c027b0a4e2cec", nil
	case "v1.15.0":
		return "medyagh/kic:v1.15.0@sha256:40d433d00a2837c8be829bd3cb0576988e377472062490bce0b18281c7f85303", nil
	case "v1.15.3":
		return "medyagh/kic:v1.15.3@sha256:f05ce52776a86c6ead806942d424de7076af3f115b0999332981a446329e6cf1", nil
	case "v1.16.1":
		return "medyagh/kic:v1.16.1@sha256:e74530d22e6a04442a97a09bdbba885ad693fcc813a0d1244da32666410d1ad1", nil
	case "v1.16.2":
		return "medyagh/kic:v1.16.2@sha256:3374a30971bf5b0011441a227fa56ef990b76125b36ca0ab8316a3c7e4f137a3", nil
	default:
		return "medyagh/kic:v1.16.2@sha256:3374a30971bf5b0011441a227fa56ef990b76125b36ca0ab8316a3c7e4f137a3", nil
	}
}
