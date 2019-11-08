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
	"strconv"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/medyagh/kic/pkg/config/cri"
	"github.com/medyagh/kic/pkg/node"
	"k8s.io/apimachinery/pkg/util/net"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/minikube/command"
)

// https://minikube.sigs.k8s.io/docs/reference/drivers/kic/
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	URL           string
	exec          command.Runner
	OciBinary     string
	ImageSha      string
	CPU           int
	Memory        int
	APIServerPort int32
}

// Config is configuration for the kic driver
type Config struct {
	MachineName   string
	CPU           int
	Memory        int
	StorePath     string
	OciBinary     string // oci tool to use (docker, podman,...)
	ImageSha      string // image name with sha to use for the node
	APIServerPort int32  // port to connect to forward from container to user's machine
}

// NewDriver returns a fully configured Kic driver
func NewDriver(c Config) *Driver {
	runner := &command.ExecRunner{}
	d := &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: c.MachineName,
			StorePath:   c.StorePath,
		},
		exec:          runner,
		OciBinary:     c.OciBinary,
		ImageSha:      c.ImageSha,
		CPU:           c.CPU,
		Memory:        c.Memory,
		APIServerPort: c.APIServerPort,
	}
	return d
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	ks := &node.Spec{ // kic spec
		Profile:           d.MachineName,
		Name:              d.MachineName + "-control-plane",
		Image:             d.ImageSha,
		CPUs:              strconv.Itoa(d.CPU),    //TODO: change kic to take int
		Memory:            strconv.Itoa(d.Memory), // TODO: change kic to take int
		Role:              "control-plane",
		ExtraMounts:       []cri.Mount{},
		ExtraPortMappings: []cri.PortMapping{},
		APIServerAddress:  "127.0.0.1", // MEDYA:TODO make configurable
		APIServerPort:     d.APIServerPort,
		IPv6:              false, // MEDYA:TODO add proxy envs here
	}
	fmt.Printf("\t(medya dbg) KickSpec: %+v\n", ks)
	// d.Node = ks.Create()
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	if d.OciBinary == "podman" {
		return "podman"
	}
	return "docker"
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
