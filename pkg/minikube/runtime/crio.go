package runtime

import (
	"github.com/docker/machine/libmachine/host"
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

// Enable idempotently enables CRIO on a host
func (r *CRIO) Enable(h *host.Host) error {
	if err := disableOthers(r, r.config); r != nil {
		glog.Warningf("disable: %v", err)
	}
	_, err := h.RunSSHCommand("sudo systemctl restart crio")
	return err
}

// Disable idempotently disables CRIO on a host
func (r *CRIO) Disable(h *host.Host) error {
	_, err := h.RunSSHCommand("sudo systemctl stop crio")
	return err
}

// KubeletOptions returns kubelet options for a runtime.
func (r *CRIO) KubeletOptions(cfg map[string]string) map[string]string {
	cfg["container-runtime"] = "remote"
	cfg["container-runtime-endpoint"] = r.SocketPath()
	cfg["image-service-endpoint"] = r.SocketPath()
	cfg["runtime-request-timeout"] = "15m"
	return cfg
}
