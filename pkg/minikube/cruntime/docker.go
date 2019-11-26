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

package cruntime

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/out"
)

// KubernetesContainerPrefix is the prefix of each kubernetes container
const KubernetesContainerPrefix = "k8s_"

// Docker contains Docker runtime state
type Docker struct {
	Socket string
	Runner CommandRunner
}

// Name is a human readable name for Docker
func (r *Docker) Name() string {
	return "Docker"
}

// Style is the console style for Docker
func (r *Docker) Style() out.StyleEnum {
	return out.Docker
}

// Version retrieves the current version of this runtime
func (r *Docker) Version() (string, error) {
	// Note: the server daemon has to be running, for this call to return successfully
	c := exec.Command("docker", "version", "--format", "'{{.Server.Version}}'")
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return "", err
	}
	return strings.Split(rr.Stdout.String(), "\n")[0], nil
}

// SocketPath returns the path to the socket file for Docker
func (r *Docker) SocketPath() string {
	return r.Socket
}

// DefaultCNI returns whether to use CNI networking by default
func (r *Docker) DefaultCNI() bool {
	return false
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *Docker) Available() error {
	_, err := exec.LookPath("docker")
	return err
}

// Active returns if docker is active on the host
func (r *Docker) Active() bool {
	c := exec.Command("systemctl", "is-active", "--quiet", "service", "docker")
	_, err := r.Runner.RunCmd(c)
	return err == nil
}

// Enable idempotently enables Docker on a host
func (r *Docker) Enable(disOthers bool) error {
	if disOthers {
		if err := disableOthers(r, r.Runner); err != nil {
			glog.Warningf("disableOthers: %v", err)
		}
	}
	c := exec.Command("sudo", "systemctl", "start", "docker")
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "enable docker.")
	}
	return nil
}

// Disable idempotently disables Docker on a host
func (r *Docker) Disable() error {
	c := exec.Command("sudo", "systemctl", "stop", "docker", "docker.socket")
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "disable docker")
	}
	return nil
}

// LoadImage loads an image into this runtime
func (r *Docker) LoadImage(path string) error {
	glog.Infof("Loading image: %s", path)
	c := exec.Command("docker", "load", "-i", path)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "loadimage docker.")
	}
	return nil

}

// KubeletOptions returns kubelet options for a runtime.
func (r *Docker) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime": "docker",
	}
}

// ListContainers returns a list of containers
func (r *Docker) ListContainers(filter string) ([]string, error) {
	filter = KubernetesContainerPrefix + filter
	rr, err := r.Runner.RunCmd(exec.Command("docker", "ps", "-a", fmt.Sprintf("--filter=name=%s", filter), "--format=\"{{.ID}}\""))
	if err != nil {
		return nil, errors.Wrapf(err, "docker ListContainers. ")
	}
	var ids []string
	for _, line := range strings.Split(rr.Stdout.String(), "\n") {
		if line != "" {
			ids = append(ids, line)
		}
	}
	return ids, nil
}

// KillContainers forcibly removes a running container based on ID
func (r *Docker) KillContainers(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	glog.Infof("Killing containers: %s", ids)
	args := append([]string{"rm", "-f"}, ids...)
	c := exec.Command("docker", args...)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "Killing containers docker.")
	}
	return nil
}

// StopContainers stops a running container based on ID
func (r *Docker) StopContainers(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	glog.Infof("Stopping containers: %s", ids)
	args := append([]string{"stop"}, ids...)
	c := exec.Command("docker", args...)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "stopping containers docker.")
	}
	return nil
}

// ContainerLogCmd returns the command to retrieve the log for a container based on ID
func (r *Docker) ContainerLogCmd(id string, len int, follow bool) string {
	var cmd strings.Builder
	cmd.WriteString("docker logs ")
	if len > 0 {
		cmd.WriteString(fmt.Sprintf("--tail %d ", len))
	}
	if follow {
		cmd.WriteString("--follow ")
	}

	cmd.WriteString(id)
	return cmd.String()
}

// SystemLogCmd returns the command to retrieve system logs
func (r *Docker) SystemLogCmd(len int) string {
	return fmt.Sprintf("sudo journalctl -u docker -n %d", len)
}
