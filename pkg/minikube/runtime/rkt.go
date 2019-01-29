package runtime

import "github.com/docker/machine/libmachine/host"

// Rkt contains Rkt runtime state
type Rkt struct {
	config Config
}

// Name is a human readable name for rkt
func (r *Rkt) Name() string {
	return "rkt"
}

// SocketPath returns the path to the socket file for rkt
func (r *Rkt) SocketPath() string {
	return "/run/Rkt/Rkt.sock"
}

// Enable idempotently enables rkt on a host
func (r *Rkt) Enable(h *host.Host) error {
	if err := disableOthers(r, r.config); r != nil {
		log.Warningf("disable: %v", err)
	}
	_, err := h.RunSSHCommand("sudo systemctl restart rkt-api rkt-metadata")
	return err
}

// Disable idempotently disables rkt on a host
func (r *Rkt) Disable(h *host.Host) error {
	_, err := h.RunSSHCommand("sudo systemctl stop rkt-api rkt-metadata")
	return err
}

// KubeletOptions returns kubelet options for a rkt
func (r *Rkt) KubeletOptions(cfg map[string]string) map[string]string {
	cfg["container-runtime"] = "rkt"
	return cfg
}
