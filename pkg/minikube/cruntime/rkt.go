package cruntime

import (
	"github.com/golang/glog"
)

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
	return ""
}

// Active returns if rkt is active on the host
func (r *Rkt) Active(cr CommandRunner) bool {
	err := cr.Run("systemctl is-active --quiet service rkt")
	if err == nil {
		return true
	}
	return false
}

// Enable idempotently enables rkt on a host
func (r *Rkt) Enable(cr CommandRunner) error {
	if err := disableOthers(r, cr); err != nil {
		glog.Warningf("disableOthers: %v", err)
	}
	return cr.Run("sudo systemctl restart rkt-api rkt-metadata")
}

// Disable idempotently disables rkt on a host
func (r *Rkt) Disable(cr CommandRunner) error {
	return cr.Run("sudo systemctl stop rkt-api rkt-metadata")
}

// LoadImage loads an image into this runtime
func (r *Rkt) LoadImage(cr CommandRunner, path string) error {
	return nil
}

// KubeletOptions returns kubelet options for a rkt
func (r *Rkt) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime": "rkt",
	}
}
