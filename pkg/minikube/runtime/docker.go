package runtime

import (
	"fmt"

	"github.com/golang/glog"
)

// Docker contains Docker runtime state
type Docker struct {
	config Config
}

// Name is a human readable name for Docker
func (r *Docker) Name() string {
	return "Docker"
}

// SocketPath returns the path to the socket file for Docker
func (r *Docker) SocketPath() string {
	return ""
}

// Active returns if docker is active on the host
func (r *Docker) Active(cr *CommandRunner) bool {
	return false
}

// Enable idempotently enables Docker on a host
func (r *Docker) Enable(cr *CommandRunner) error {
	if err := disableOthers(r, cr); r != nil {
		glog.Warningf("disable: %v", err)
	}
	return cr.Run("sudo systemctl restart docker")
	return err
}

// Disable idempotently disables Docker on a host
func (r *Docker) Disable(cr *CommandRunner) error {
	return cr.Run("sudo systemctl stop docker")
	return err
}

// LoadImage loads an image into this runtime
func (r *Docker) LoadImage(cr *CommandRunner, path string) error {
	return cr.Run(fmt.Sprintf("docker load -i %s", path))
	return err
}

// KubeletOptions returns kubelet options for a runtime.
func (r *Docker) KubeletOptions(cfg map[string]string) map[string]string {
	cfg["container-runtime"] = "docker"
	return cfg
}
