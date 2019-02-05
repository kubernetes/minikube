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

// SocketPath returns the path to the socket file for Docker
func (r *Docker) SocketPath() string {
	return r.Socket
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
	return r.Runner.Run("sudo systemctl restart docker")
}

// Disable idempotently disables Docker on a host
func (r *Docker) Disable() error {
	return r.Runner.Run("sudo systemctl stop docker docker.socket")
}

// LoadImage loads an image into this runtime
func (r *Docker) LoadImage(path string) error {
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
	return strings.Split(content, "\n"), nil
}

// KillContainers forcibly removes a running pod based on ID
func (r *Docker) KillContainers(ids []string) error {
	return r.Runner.Run(fmt.Sprintf("docker rm -f %s", strings.Join(ids, " ")))
}

// StopContainers stops a running pod based on ID
func (r *Docker) StopContainers(ids []string) error {
	return r.Runner.Run(fmt.Sprintf("docker stop %s", strings.Join(ids, " ")))
}
