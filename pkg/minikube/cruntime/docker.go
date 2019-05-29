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
)

// Docker contains Docker runtime state
type Docker struct {
	Socket string
	Runner CommandRunner
}

// Name is a human readable name for Docker
func (r *Docker) Name() string {
	return "Docker"
}

// Version retrieves the current version of this runtime
func (r *Docker) Version() (string, error) {
	// Note: the server daemon has to be running, for this call to return successfully
	ver, err := r.Runner.CombinedOutput("docker version --format '{{.Server.Version}}'")
	if err != nil {
		return "", err
	}

	return strings.Split(ver, "\n")[0], nil
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
	err := r.Runner.Run("systemctl is-active --quiet service docker")
	return err == nil
}

// Enable idempotently enables Docker on a host
func (r *Docker) Enable() error {
	if err := disableOthers(r, r.Runner); err != nil {
		glog.Warningf("disableOthers: %v", err)
	}
	return r.Runner.Run("sudo systemctl start docker")
}

// Disable idempotently disables Docker on a host
func (r *Docker) Disable() error {
	return r.Runner.Run("sudo systemctl stop docker docker.socket")
}

// LoadImage loads an image into this runtime
func (r *Docker) LoadImage(path string) error {
	glog.Infof("Loading image: %s", path)
	return r.Runner.Run(fmt.Sprintf("docker load -i %s", path))
}

// KubeletOptions returns kubelet options for a runtime.
func (r *Docker) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime": "docker",
	}
}

// ListContainers returns a list of containers
func (r *Docker) ListContainers(filter string) ([]string, error) {
	content, err := r.Runner.CombinedOutput(fmt.Sprintf(`docker ps -a --filter="name=%s" --format="{{.ID}}"`, filter))
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, line := range strings.Split(content, "\n") {
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
	return r.Runner.Run(fmt.Sprintf("docker rm -f %s", strings.Join(ids, " ")))
}

// StopContainers stops a running container based on ID
func (r *Docker) StopContainers(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	glog.Infof("Stopping containers: %s", ids)
	return r.Runner.Run(fmt.Sprintf("docker stop %s", strings.Join(ids, " ")))
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
