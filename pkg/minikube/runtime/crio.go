package runtime

import (
	"fmt"

	"github.com/golang/glog"
)

// CRIO contains CRIO runtime state
type CRIO struct {
	config Config
}

// Name is a human readable name for CRIO
func (r *CRIO) Name() string {
	return "CRIO"
}

// SocketPath returns the path to the socket file for CRIO
func (r *CRIO) SocketPath() string {
	return "/var/run/crio/crio.sock"
}

// Active returns if CRIO is active on the host
func (r *CRIO) Active(cr CommandRunner) bool {
	return false
}

// Enable idempotently enables CRIO on a host
func (r *CRIO) Enable(cr CommandRunner) error {
	if err := disableOthers(r, cr); r != nil {
		glog.Warningf("disable: %v", err)
	}
	return cr.Run("sudo systemctl restart crio")
}

// Disable idempotently disables CRIO on a host
func (r *CRIO) Disable(cr CommandRunner) error {
	return cr.Run("sudo systemctl stop crio")
}

// LoadImage loads an image into this runtime
func (r *CRIO) LoadImage(cr CommandRunner, path string) error {
	return cr.Run(fmt.Sprintf("sudo podman load -i %s", path))
}

// KubeletOptions returns kubelet options for a runtime.
func (r *CRIO) KubeletOptions(cfg map[string]string) map[string]string {
	cfg["container-runtime"] = "remote"
	cfg["container-runtime-endpoint"] = r.SocketPath()
	cfg["image-service-endpoint"] = r.SocketPath()
	cfg["runtime-request-timeout"] = "15m"
	return cfg
}
