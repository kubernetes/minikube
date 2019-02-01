package cruntime

import (
	"fmt"
	"os/exec"

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
	err := cr.Run("systemctl is-active --quiet service containerd")
	return err == nil
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *Containerd) Available(CommandRunner) error {
	_, err := exec.LookPath("containerd")
	return err
}

// Enable idempotently enables containerd on a host
func (r *Containerd) Enable(cr CommandRunner) error {
	if err := disableOthers(r, cr); err != nil {
		glog.Warningf("disableOthers: %v", err)
	}
	if err := enableIPForwarding(cr); err != nil {
		return err
	}
	return cr.Run("sudo systemctl start containerd")
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
func (r *Containerd) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime":          "remote",
		"container-runtime-endpoint": fmt.Sprintf("unix://%s", r.SocketPath()),
		"image-service-endpoint":     fmt.Sprintf("unix://%s", r.SocketPath()),
		"runtime-request-timeout":    "15m",
	}
}

// ListContainers returns a list of managed by this container runtime
func (r *Containerd) ListContainers(cr CommandRunner, filter string) ([]string, error) {
	return listCRIContainers(cr, filter)
}

// KillContainers removes containers based on ID
func (r *Containerd) KillContainers(cr CommandRunner, ids []string) error {
	return killCRIContainers(cr, ids)
}

// StopContainers stops containers based on ID
func (r *Containerd) StopContainers(cr CommandRunner, ids []string) error {
	return stopCRIContainers(cr, ids)
}
