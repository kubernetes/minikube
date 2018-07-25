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

	"github.com/golang/glog"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/net"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
)

const driverName = "none"

const dockerList = `docker ps -a --filter="name=k8s_" -q`
const dockerKill = ` | xargs docker kill`
const dockerStop = ` | xargs docker stop`
const dockerRm = ` | xargs docker rm`

// none Driver is a driver designed to run kubeadm w/o a VM
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

// PreCreateCheck checks for correct privileges and dependencies
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
	ip, err := net.ChooseBindAddress(nil)
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "", fmt.Errorf("driver does not support ssh commands")
}

func (d *Driver) GetSSHPort() (int, error) {
	return 0, fmt.Errorf("driver does not support ssh commands")
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetState() (state.State, error) {
	const statusCmd = `systemctl is-active kubelet || true` // is-active returns status via exit-code, we don't need that

	out, err := runCommand(statusCmd, true)
	if err != nil {
		return state.None, err
	}

	out = strings.TrimSpace(out)
	if "active" == out {
		return state.Running, nil
	} else if "inactive" == out {
		return state.Stopped, nil
	} else {
		return state.None, fmt.Errorf("Invalid `systemctl is-active` output: %s", out)
	}
}

func (d *Driver) Kill() error {
	var killCmds = []string{
		`systemctl kill kubelet.service`,
		dockerList + dockerKill + dockerRm,
	}

	for _, cmdStr := range killCmds {
		if _, err := runCommand(cmdStr, true); err != nil {
			glog.Warningf("Failed to run command `%s`: %s", cmdStr, err)
		}
	}
	return nil
}

func (d *Driver) Remove() error {
	var rmCmds = []string{
		`systemctl stop kubelet.service`,
		dockerList + dockerKill + dockerRm,
		`rm -rf /data/minikube`,
		`rm -rf /etc/kubernetes/manifests`,
		`rm -rf /var/lib/minikube`,
	}

	for _, cmdStr := range rmCmds {
		if _, err := runCommand(cmdStr, true); err != nil {
			glog.Warningf("Failed to run command `%s`: %s", cmdStr, err)
		}
	}
	return nil
}

func (d *Driver) Restart() error {
	var restartCmds = []string{
		`systemctl stop kubelet.service`,
		dockerList + dockerStop + dockerRm,
		`systemctl start kubelet.service`,
	}

	for _, cmdStr := range restartCmds {
		if _, err := runCommand(cmdStr, true); err != nil {
			glog.Warningf("Failed to run command `%s`: %s", cmdStr, err)
		}
	}
	return nil
}

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

func (d *Driver) Stop() error {
	var stopCmds = []string{
		`systemctl stop kubelet.service`,
		dockerList + dockerStop + dockerRm,
	}

	for _, cmdStr := range stopCmds {
		if _, err := runCommand(cmdStr, true); err != nil {
			glog.Warningf("Failed to run command `%s`: %s", cmdStr, err)
		}
	}
	return nil
}

func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("driver does not support ssh commands")
}

func runCommand(command string, sudo bool) (string, error) {
	var cmd *exec.Cmd
	if sudo {
		cmd = exec.Command("sudo", "/bin/bash", "-c", command)
	} else {
		cmd = exec.Command("/bin/bash", "-c", command)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), errors.Wrap(err, stderr.String())
	}
	return stdout.String(), nil
}
