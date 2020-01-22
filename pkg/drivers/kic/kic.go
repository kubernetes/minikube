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
	"k8s.io/minikube/pkg/minikube/constants"
)

// DefaultPodCIDR is The CIDR to be used for pods inside the node.
const DefaultPodCIDR = "10.244.0.0/16"

// DefaultBindIPV4 is The default IP the container will bind to.
const DefaultBindIPV4 = "127.0.0.1"

// BaseImage is the base image is used to spin up kic containers created by kind.
const BaseImage = "gcr.io/k8s-minikube/kicbase:v0.0.1@sha256:c4ad2938877d2ae0d5b7248a5e7182ff58c0603165c3bedfe9d503e2d380a0db"

// OverlayImage is the cni plugin used for overlay image, created by kind.
const OverlayImage = "kindest/kindnetd:0.5.3"

// Driver represents a kic driver https://minikube.sigs.k8s.io/docs/reference/drivers/kic/
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
	MachineName  string            // maps to the container name being created
	CPU          int               // Number of CPU cores assigned to the container
	Memory       int               // max memory in MB
	StorePath    string            // libmachine store path
	OCIBinary    string            // oci tool to use (docker, podman,...)
	ImageDigest  string            // image name with sha to use for the node
	HostBindPort int               // port to connect to forward from container to user's machine
	Mounts       []oci.Mount       // mounts
	PortMappings []oci.PortMapping // container port mappings
	Envs         map[string]string // key,value of environment variables passed to the node
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
		OCIBinary:  c.OCIBinary,
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
		ExtraArgs:    []string{"--expose", fmt.Sprintf("%d", d.NodeConfig.HostBindPort)},
		OCIBinary:    d.NodeConfig.OCIBinary,
	}

	// control plane specific options
	params.PortMappings = append(params.PortMappings, oci.PortMapping{
		ListenAddress: "127.0.0.1",
		HostPort:      int32(d.NodeConfig.HostBindPort),
		ContainerPort: constants.APIServerPort,
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
	switch o {
	case "running":
		return state.Running, nil
	case "exited":
		return state.Stopped, nil
	case "paused":
		return state.Paused, nil
	case "restarting":
		return state.Starting, nil
	case "dead":
		return state.Error, nil
	default:
		return state.None, fmt.Errorf("unknown state")
	}
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
	cmd := exec.Command(d.NodeConfig.OCIBinary, "unpause", d.MachineName)
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
	// TODO:medyagh maybe make it idempotent
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
