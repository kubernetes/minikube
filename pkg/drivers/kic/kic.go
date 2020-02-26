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
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
)

// Driver represents a kic driver https://minikube.sigs.k8s.io/docs/reference/drivers/docker
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	URL        string
	exec       command.Runner
	NodeConfig Config
	OCIBinary  string // docker,podman
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
	params := oci.CreateParams{
		Name:          d.NodeConfig.MachineName,
		Image:         d.NodeConfig.ImageDigest,
		ClusterLabel:  oci.ProfileLabelKey + "=" + d.MachineName,
		CPUs:          strconv.Itoa(d.NodeConfig.CPU),
		Memory:        strconv.Itoa(d.NodeConfig.Memory) + "mb",
		Envs:          d.NodeConfig.Envs,
		ExtraArgs:     []string{"--expose", fmt.Sprintf("%d", d.NodeConfig.APIServerPort)},
		OCIBinary:     d.NodeConfig.OCIBinary,
		APIServerPort: d.NodeConfig.APIServerPort,
	}

	// control plane specific options
	params.PortMappings = append(params.PortMappings, oci.PortMapping{
		ListenAddress: oci.DefaultBindIPV4,
		ContainerPort: constants.APIServerPort,
	},
		oci.PortMapping{
			ListenAddress: oci.DefaultBindIPV4,
			ContainerPort: constants.SSHPort,
		},
		oci.PortMapping{
			ListenAddress: oci.DefaultBindIPV4,
			ContainerPort: constants.DockerDaemonPort,
		},
	)
	err := oci.CreateContainerNode(params)
	if err != nil {
		return errors.Wrap(err, "create kic node")
	}

	if err := d.prepareSSH(); err != nil {
		return errors.Wrap(err, "prepare kic ssh")
	}
	return nil
}

// prepareSSH will generate keys and copy to the container so minikube ssh works
func (d *Driver) prepareSSH() error {
	keyPath := d.GetSSHKeyPath()
	glog.Infof("Creating ssh key for kic: %s...", keyPath)
	if err := ssh.GenerateSSHKey(keyPath); err != nil {
		return errors.Wrap(err, "generate ssh key")
	}

	cmder := command.NewKICRunner(d.NodeConfig.MachineName, d.NodeConfig.OCIBinary)
	f, err := assets.NewFileAsset(d.GetSSHKeyPath()+".pub", "/home/docker/.ssh/", "authorized_keys", "0644")
	if err != nil {
		return errors.Wrap(err, "create pubkey assetfile ")
	}
	if err := cmder.Copy(f); err != nil {
		return errors.Wrap(err, "copying pub key")
	}
	if rr, err := cmder.RunCmd(exec.Command("chown", "docker:docker", "/home/docker/.ssh/authorized_keys")); err != nil {
		return errors.Wrapf(err, "apply authorized_keys file ownership, output %s", rr.Output())
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
	ip, _, err := oci.ContainerIPs(d.OCIBinary, d.MachineName)
	return ip, err
}

// GetExternalIP returns an IP which is accissble from outside
func (d *Driver) GetExternalIP() (string, error) {
	return oci.DefaultBindIPV4, nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return oci.DefaultBindIPV4, nil
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	p, err := oci.HostPortBinding(d.OCIBinary, d.MachineName, constants.SSHPort)
	if err != nil {
		return p, errors.Wrap(err, "get ssh host-port")
	}
	return p, nil
}

// GetSSHUsername returns the ssh username
func (d *Driver) GetSSHUsername() string {
	return "docker"
}

// GetSSHKeyPath returns the ssh key path
func (d *Driver) GetSSHKeyPath() string {
	if d.SSHKeyPath == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

// GetURL returns ip of the container running kic control-panel
func (d *Driver) GetURL() (string, error) {
	p, err := oci.HostPortBinding(d.NodeConfig.OCIBinary, d.MachineName, d.NodeConfig.APIServerPort)
	url := fmt.Sprintf("https://%s", net.JoinHostPort("127.0.0.1", fmt.Sprint(p)))
	if err != nil {
		return url, errors.Wrap(err, "api host port binding")
	}
	return url, nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	if err := oci.PointToHostDockerDaemon(); err != nil {
		return state.Error, errors.Wrap(err, "point host docker daemon")
	}
	// allow no more than 2 seconds for this. when this takes long this means deadline passed
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, d.NodeConfig.OCIBinary, "inspect", "-f", "{{.State.Status}}", d.MachineName)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		glog.Errorf("GetState for %s took longer than normal. Restarting your %s daemon might fix this issue.", d.MachineName, d.OCIBinary)
		return state.Error, fmt.Errorf("inspect %s timeout", d.MachineName)
	}
	o := strings.TrimSpace(string(out))
	if err != nil {
		return state.Error, errors.Wrapf(err, "get container %s status", d.MachineName)
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
	if _, err := oci.ContainerID(d.OCIBinary, d.MachineName); err != nil {
		log.Warnf("could not find the container %s to remove it.", d.MachineName)
	}
	cmd := exec.Command(d.NodeConfig.OCIBinary, "rm", "-f", "-v", d.MachineName)
	o, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(o))
	if err != nil {
		if strings.Contains(out, "is already in progress") {
			log.Warnf("Docker engine is stuck. please restart docker daemon on your computer.", d.MachineName)
		}
		return errors.Wrapf(err, "removing container %s, output %s", d.MachineName, out)
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
