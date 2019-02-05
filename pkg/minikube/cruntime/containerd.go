package cruntime

import (
	"fmt"

	"github.com/golang/glog"
)

// Containerd contains containerd runtime state
type Containerd struct {
	Socket string
	Runner CommandRunner
}

// Name is a human readable name for containerd
func (r *Containerd) Name() string {
	return "containerd"
}

// SocketPath returns the path to the socket file for containerd
func (r *Containerd) SocketPath() string {
	if r.Socket != "" {
		return r.Socket
	}
	return "/run/containerd/containerd.sock"
}

// Active returns if containerd is active on the host
func (r *Containerd) Active() bool {
	err := r.Runner.Run("systemctl is-active --quiet service containerd")
	return err == nil
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *Containerd) Available() error {
	return r.Runner.Run("command -v crio")
}

// Enable idempotently enables containerd on a host
func (r *Containerd) Enable() error {
	if err := disableOthers(r, r.Runner); err != nil {
		glog.Warningf("disableOthers: %v", err)
	}
	if err := populateCRIConfig(r.Runner, r.SocketPath()); err != nil {
		return err
	}
	if err := enableIPForwarding(r.Runner); err != nil {
		return err
	}
	// Oherwise, containerd will fail API requests with 'Unimplemented'
	return r.Runner.Run("sudo systemctl restart containerd")
}

// Disable idempotently disables containerd on a host
func (r *Containerd) Disable() error {
	return r.Runner.Run("sudo systemctl stop containerd")
}

// LoadImage loads an image into this runtime
func (r *Containerd) LoadImage(path string) error {
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
func (r *Containerd) ListContainers(filter string) ([]string, error) {
	return listCRIContainers(r.Runner, filter)
}

// KillContainers removes containers based on ID
func (r *Containerd) KillContainers(ids []string) error {
	return killCRIContainers(r.Runner, ids)
}

// StopContainers stops containers based on ID
func (r *Containerd) StopContainers(ids []string) error {
	return stopCRIContainers(r.Runner, ids)
}
