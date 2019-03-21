/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cruntime

import (
	"fmt"
	"strings"

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

// Version retrieves the current version of this runtime
func (r *Containerd) Version() (string, error) {
	ver, err := r.Runner.CombinedOutput("containerd --version")
	if err != nil {
		return "", err
	}

	// containerd github.com/containerd/containerd v1.2.0 c4446665cb9c30056f4998ed953e6d4ff22c7c39
	words := strings.Split(ver, " ")
	if len(words) >= 4 && words[0] == "containerd" {
		return strings.Replace(words[2], "v", "", 1), nil
	}
	return "", fmt.Errorf("unknown version: %q", ver)
}

// SocketPath returns the path to the socket file for containerd
func (r *Containerd) SocketPath() string {
	if r.Socket != "" {
		return r.Socket
	}
	return "/run/containerd/containerd.sock"
}

// DefaultCNI returns whether to use CNI networking by default
func (r *Containerd) DefaultCNI() bool {
	return true
}

// Active returns if containerd is active on the host
func (r *Containerd) Active() bool {
	err := r.Runner.Run("systemctl is-active --quiet service containerd")
	return err == nil
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *Containerd) Available() error {
	return r.Runner.Run("command -v containerd")
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
	glog.Infof("Loading image: %s", path)
	return r.Runner.Run(fmt.Sprintf("sudo ctr images import %s", path))
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

// ContainerLogCmd returns the command to retrieve the log for a container based on ID
func (r *Containerd) ContainerLogCmd(id string, len int, follow bool) string {
	return criContainerLogCmd(id, len, follow)
}
