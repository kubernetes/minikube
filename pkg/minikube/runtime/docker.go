package runtime

import (
	"github.com/docker/machine/libmachine/host"
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

// Enable idempotently enables Docker on a host
func (r *Docker) Enable(h *host.Host) error {
	if err := disableOthers(r, r.config); r != nil {
		glog.Warningf("disable: %v", err)
	}
	_, err := h.RunSSHCommand("sudo systemctl restart docker")
	return err
}

// Disable idempotently disables Docker on a host
func (r *Docker) Disable(h *host.Host) error {
	_, err := h.RunSSHCommand("sudo systemctl stop docker")
	return err
}

// KubeletOptions returns kubelet options for a runtime.
func (r *Docker) KubeletOptions(cfg map[string]string) map[string]string {
	cfg["container-runtime"] = "remote"
	cfg["container-runtime-endpoint"] = socket
	cfg["image-service-endpoint"] = socket
	cfg["runtime-request-timeout"] = "15m"
	return cfg
}
