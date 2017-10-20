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
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/minikube/constants"
)

const driverName = "none"
const dockerkillcmd = `docker rm $(docker kill $(docker ps -a --filter="name=k8s_" --format="{{.ID}}"))`
const dockerstopcmd = `docker stop $(docker ps -a --filter="name=k8s_" --format="{{.ID}}")`

// none Driver is a driver designed to run localkube w/o a VM
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
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

// PreCreateCheck checks for correct priviledges and dependencies
func (d *Driver) PreCreateCheck() error {
	// check that docker is on path
	_, err := exec.LookPath("docker")
	if err != nil {
		return errors.Wrap(err, "docker cannot be found on the path for this machine. "+
			"A docker installation is a requirement for using the none driver")
	}

	return nil
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

func (d *Driver) GetSSHPort() (int, error) {
	return 0, fmt.Errorf("driver does not support ssh commands")
}

func (d *Driver) GetURL() (string, error) {
	return "tcp://127.0.0.1:2376", nil
}

func (d *Driver) GetState() (state.State, error) {
	var statuscmd = fmt.Sprintf("if [[ `systemctl` =~ -\\.mount ]] &>/dev/null; "+`then
  sudo systemctl is-active localkube &>/dev/null && echo "Running" || echo "Stopped"
else
  if ps $(cat %s) &>/dev/null; then
    echo "Running"
  else
    echo "Stopped"
  fi
fi
`, constants.LocalkubePIDPath)

	out, err := runCommand(statuscmd, true)
	if err != nil {
		return state.None, err
	}
	s := strings.TrimSpace(out)
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
		return errors.Wrap(err, "stopping the localkube service")
	}
	cmd = exec.Command("sudo", "rm", "-rf", "/var/lib/localkube")
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "removing localkube")
	}
	return nil
}

func (d *Driver) Remove() error {
	cmd := exec.Command("sudo", "systemctl", "stop", "localkube.service")
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "stopping localkube service")
	}
	cmd = exec.Command("sudo", "rm", "-rf", "/var/lib/localkube")
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "removing localkube")

	}
	runCommand(dockerkillcmd, false)

	return nil
}

func (d *Driver) Restart() error {
	cmd := exec.Command("sudo", "systemctl", "restart", "localkube.service")
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Start() error {
	d.IPAddress = "127.0.0.1"
	d.URL = "127.0.0.1:8080"
	return nil
}

func (d *Driver) Stop() error {
	var stopcmd = fmt.Sprintf("if [[ `systemctl` =~ -\\.mount ]] &>/dev/null; "+`then
  sudo systemctl stop localkube.service
else
	sudo kill $(cat %s)
fi
`, constants.LocalkubePIDPath)
	_, err := runCommand(stopcmd, false)
	if err != nil {
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
	runCommand(dockerstopcmd, false)
	return nil
}

func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("driver does not support ssh commands")
}

func runCommand(command string, sudo bool) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	if sudo {
		cmd = exec.Command("sudo", "/bin/bash", "-c", command)
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, stderr.String())
	}
	return out.String(), nil
}
