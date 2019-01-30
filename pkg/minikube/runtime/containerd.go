package runtime

import (
	"fmt"

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

// Active returns if containerd is active on the host
func (r *Containerd) Active(cr CommandRunner) bool {
	return false
}

// Enable idempotently enables containerd on a host
func (r *Containerd) Enable(cr CommandRunner) error {
	if err := disableOthers(r, cr); r != nil {
		glog.Warningf("disable: %v", err)
	}
	return cr.Run("sudo systemctl restart containerd")
}

// Disable idempotently disables containerd on a host
func (r *Containerd) Disable(cr CommandRunner) error {
	return cr.Run("sudo systemctl stop containerd")
}

// LoadImage loads an image into this runtime
func (r *Containerd) LoadImage(cr CommandRunner, path string) error {
	return nil
}

// KubeletOptions returns kubelet options for a containerd
func (r *Containerd) KubeletOptions(cfg map[string]string) map[string]string {
	cfg["container-runtime"] = "remote"
	cfg["container-runtime-endpoint"] = fmt.Sprintf("unix://%s", r.SocketPath())
	cfg["image-service-endpoint"] = fmt.Sprintf("unix://%s", r.SocketPath())
	cfg["runtime-request-timeout"] = "15m"
	return cfg
}
