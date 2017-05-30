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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/constants"
)

const driverName = "none"

// none Driver is a driver designed to run localkube w/o a VM
type Driver struct {
	*drivers.BaseDriver
	URL string
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// PreCreateCheck checks that VBoxManage exists and works
func (d *Driver) PreCreateCheck() error {
	// check that systemd is installed as it is a requirement
	if _, err := exec.LookPath("systemctl"); err != nil {
		return errors.New("systemd is a requirement in order to use the none driver")
	}
	return nil
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{}
}

func (d *Driver) Create() error {
	// creation for the none driver is handled by commands.go
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetIP() (string, error) {
	return "127.0.0.1", nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "", fmt.Errorf("driver does not support ssh commands")
}

func (d *Driver) GetSSHKeyPath() string {
	return ""
}

func (d *Driver) GetSSHPort() (int, error) {
	return 0, fmt.Errorf("driver does not support ssh commands")
}

func (d *Driver) GetSSHUsername() string {
	return ""
}

func (d *Driver) GetURL() (string, error) {
	return "127.0.0.1:8080", nil
}

func (d *Driver) GetState() (state.State, error) {
	command := `sudo systemctl is-active localkube 2>&1 1>/dev/null && echo "Running" || echo "Stopped"`

	path := filepath.Join(constants.GetMinipath(), "tmp-cmd")
	ioutil.WriteFile(filepath.Join(constants.GetMinipath(), "tmp-cmd"), []byte(command), os.FileMode(0644))
	defer os.Remove(path)
	cmd := exec.Command("sudo", "/bin/sh", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return state.None, err
	}
	s := strings.TrimSpace(string(out))
	if state.Running.String() == s {
		return state.Running, nil
	} else if state.Stopped.String() == s {
		return state.Stopped, nil
	} else {
		return state.None, fmt.Errorf("Error: Unrecognize output from GetLocalkubeStatus: %s", s)
	}
}

func (d *Driver) Kill() error {
	cmd := exec.Command("sudo", "systemctl", "stop", "localkube.service")
	if err := cmd.Start(); err != nil {
		return err
	}
	cmd = exec.Command("sudo", "rm", "-rf", "/var/lib/localkube")
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Remove() error {
	cmd := exec.Command("sudo", "systemctl", "stop", "localkube.service")
	if err := cmd.Start(); err != nil {
		return err
	}
	cmd = exec.Command("sudo", "rm", "-rf", "/var/lib/localkube")
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Restart() error {
	cmd := exec.Command("sudo", "systemctl", "restart", "localkube.service")
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func (d *Driver) Start() error {
	d.IPAddress = "127.0.0.1"
	d.URL = "127.0.0.1:8080"
	return nil
}

func (d *Driver) Stop() error {
	cmd := exec.Command("sudo", "systemctl", "stop", "localkube.service")
	if err := cmd.Start(); err != nil {
		return err
	}

	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}
		if s != state.Running {
			break
		}
	}
	return nil
}

func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("driver does not support ssh commands")
}
