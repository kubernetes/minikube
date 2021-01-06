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

package generic

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/klog/v2"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	EnginePort int
	SSHKey     string
}

const (
	defaultTimeout = 15 * time.Second
)

// NewDriver creates and returns a new instance of the driver
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		EnginePort: engine.DefaultPort,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "generic"
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUser
}

func (d *Driver) GetSSHKeyPath() string {
	return d.SSHKeyPath
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.EnginePort = flags.Int("generic-engine-port")
	d.IPAddress = flags.String("generic-ip-address")
	d.SSHUser = flags.String("generic-ssh-user")
	d.SSHKey = flags.String("generic-ssh-key")
	d.SSHPort = flags.Int("generic-ssh-port")

	if d.IPAddress == "" {
		return errors.New("generic driver requires the --generic-ip-address option")
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if d.SSHKey != "" {
		if _, err := os.Stat(d.SSHKey); os.IsNotExist(err) {
			return fmt.Errorf("SSH key does not exist: %q", d.SSHKey)
		}

		// TODO: validate the key is a valid key
	}

	return nil
}

func (d *Driver) Create() error {
	if d.SSHKey == "" {
		log.Info("No SSH key specified. Assuming an existing key at the default location.")
	} else {
		log.Info("Importing SSH key...")

		d.SSHKeyPath = d.ResolveStorePath(path.Base(d.SSHKey))
		if err := copySSHKey(d.SSHKey, d.SSHKeyPath); err != nil {
			return err
		}

		if err := copySSHKey(d.SSHKey+".pub", d.SSHKeyPath+".pub"); err != nil {
			log.Infof("Couldn't copy SSH public key : %s", err)
		}
	}

	log.Debugf("IP: %s", d.IPAddress)

	return nil
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(d.EnginePort))), nil
}

func (d *Driver) GetState() (state.State, error) {
	address := net.JoinHostPort(d.IPAddress, strconv.Itoa(d.SSHPort))

	_, err := net.DialTimeout("tcp", address, defaultTimeout)
	if err != nil {
		return state.Stopped, nil
	}

	return state.Running, nil
}

// Start a host
func (d *Driver) Start() error {
	return nil
}

// Stop a host gracefully, including any containers that we are managing.
func (d *Driver) Stop() error {
	exec := command.NewSSHRunner(d)
	if err := sysinit.New(exec).Stop("kubelet"); err != nil {
		klog.Warningf("couldn't stop kubelet. will continue with stop anyways: %v", err)
		if err := sysinit.New(exec).ForceStop("kubelet"); err != nil {
			klog.Warningf("couldn't force stop kubelet. will continue with stop anyways: %v", err)
		}
	}
	/* TODO
	containers, err := d.runtime.ListContainers(cruntime.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "containers")
	}
	if len(containers) > 0 {
		if err := d.runtime.StopContainers(containers); err != nil {
			return errors.Wrap(err, "stop containers")
		}
	}
	*/
	klog.Infof("generic driver is stopped!")
	return nil
}

// Restart a host
func (d *Driver) Restart() error {
	exec := command.NewSSHRunner(d)
	return restartKubelet(exec)
}

// Kill stops a host forcefully, including any containers that we are managing.
func (d *Driver) Kill() error {
	exec := command.NewSSHRunner(d)
	if err := sysinit.New(exec).ForceStop("kubelet"); err != nil {
		klog.Warningf("couldn't force stop kubelet. will continue with kill anyways: %v", err)
	}

	/* TODO
	// First try to gracefully stop containers
	containers, err := d.runtime.ListContainers(cruntime.ListOptions{})
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

	containers, err = d.runtime.ListContainers(cruntime.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "containers")
	}
	if len(containers) == 0 {
		return nil
	}
	if err := d.runtime.KillContainers(containers); err != nil {
		return errors.Wrap(err, "kill")
	}
	*/
	return nil

}

func (d *Driver) Remove() error {
	return nil
}

func copySSHKey(src, dst string) error {
	if err := mcnutils.CopyFile(src, dst); err != nil {
		return fmt.Errorf("unable to copy ssh key: %s", err)
	}

	if err := os.Chmod(dst, 0600); err != nil {
		return fmt.Errorf("unable to set permissions on the ssh key: %s", err)
	}

	return nil
}

// restartKubelet restarts the kubelet
func restartKubelet(cr command.Runner) error {
	klog.Infof("restarting kubelet.service ...")
	c := exec.Command("sudo", "systemctl", "restart", "kubelet.service")
	if _, err := cr.RunCmd(c); err != nil {
		return err
	}
	return nil
}
