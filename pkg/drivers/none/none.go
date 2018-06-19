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
const dockerstopcmd = `docker kill $(docker ps -a --filter="name=k8s_" --format="{{.ID}}")`

var dockerkillcmd = fmt.Sprintf(`docker rm $(%s)`, dockerstopcmd)

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
	var statuscmd = fmt.Sprintf(
		`sudo systemctl is-active kubelet &>/dev/null && echo "Running" || echo "Stopped"`)

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
		return state.None, fmt.Errorf("Error: Unrecognize output from GetState: %s", s)
	}
}

func (d *Driver) Kill() error {
	for _, cmdStr := range [][]string{
		{"systemctl", "stop", "kubelet.service"},
		{"rm", "-rf", "/var/lib/minikube"},
	} {
		cmd := exec.Command("sudo", cmdStr...)
		if out, err := cmd.CombinedOutput(); err != nil {
			glog.Warningf("Error %s running command: %s. Output: %s", err, cmdStr, string(out))
		}
	}
	return nil
}

func (d *Driver) Remove() error {
	rmCmd := `sudo systemctl stop kubelet.service
	sudo rm -rf /data/minikube
	sudo rm -rf /etc/kubernetes/manifests
	sudo rm -rf /var/lib/minikube || true`

	for _, cmdStr := range []string{rmCmd, dockerkillcmd} {
		if out, err := runCommand(cmdStr, true); err != nil {
			glog.Warningf("Error %s running command: %s, Output: %s", err, cmdStr, out)
		}
	}

	return nil
}

func (d *Driver) Restart() error {
	restartCmd := `
	if systemctl is-active kubelet.service; then
		sudo systemctl restart kubelet.service
	fi`

	cmd := exec.Command(restartCmd)
	if err := cmd.Start(); err != nil {
		return err
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
	var stopcmd = fmt.Sprintf("if [[ `systemctl` =~ -\\.mount ]] &>/dev/null; " + `then
for svc in "kubelet"; do
	sudo systemctl stop "$svc".service || true
done
fi
`)
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
	if out, err := runCommand(dockerstopcmd, false); err != nil {
		glog.Warningf("Error %s running command %s. Output: %s", err, dockerstopcmd, out)
	}
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
