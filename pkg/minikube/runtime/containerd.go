package runtime

import (
	"fmt"

	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
)

// Containerd contains containerd runtime state
type Containerd struct {
	config Config
}

// Name is a human readable name for containerd
func (r *Containerd) Name() string {
	return "containerd"
}

// SocketPath returns the path to the socket file for containerd
func (r *Containerd) SocketPath() string {
	return "/run/containerd/containerd.sock"
}

// Enable idempotently enables containerd on a host
func (r *Containerd) Enable(h *host.Host) error {
	if err := disableOthers(r, r.config); r != nil {
		glog.Warningf("disable: %v", err)
	}
	_, err := h.RunSSHCommand("sudo systemctl restart containerd")
	return err
}

// Disable idempotently disables containerd on a host
func (r *Containerd) Disable(h *host.Host) error {
	_, err := h.RunSSHCommand("sudo systemctl stop containerd")
	return err
}

// KubeletOptions returns kubelet options for a containerd
func (r *Containerd) KubeletOptions(cfg map[string]string) map[string]string {
	cfg["container-runtime"] = "remote"
	cfg["container-runtime-endpoint"] = fmt.Sprintf("unix://%s", r.SocketPath())
	cfg["image-service-endpoint"] = fmt.Sprintf("unix://%s", r.SocketPath())
	cfg["runtime-request-timeout"] = "15m"
	return cfg
}
