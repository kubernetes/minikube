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

	"github.com/golang/glog"
)

// Rkt contains rkt runtime state
type Rkt struct {
	Socket string
	Runner CommandRunner
}

// Name is a human readable name for rkt
func (r *Rkt) Name() string {
	return "rkt"
}

// SocketPath returns the path to the socket file for rkt/rktlet
func (r *Rkt) SocketPath() string {
	if r.Socket != "" {
		return r.Socket
	}
	return "/var/run/rktlet.sock"
}

// DefaultCNI returns whether to use CNI networking by default
func (r *Rkt) DefaultCNI() bool {
	return true
}

// Active returns if rkt is active on the host
func (r *Rkt) Active() bool {
	err := r.Runner.Run("systemctl is-active --quiet service rkt-api")
	return err == nil
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *Rkt) Available() error {
	return r.Runner.Run("command -v rkt")
}

// Enable idempotently enables rkt on a host
func (r *Rkt) Enable() error {
	if err := disableOthers(r, r.Runner); err != nil {
		glog.Warningf("disableOthers: %v", err)
	}
	if err := populateCRIConfig(r.Runner, r.SocketPath()); err != nil {
		return err
	}
	if err := enableIPForwarding(r.Runner); err != nil {
		return err
	}
	err := r.Runner.Run("sudo systemctl start rkt-api")
	if err == nil {
		err = r.Runner.Run("sudo systemctl start rkt-metadata")
	}
	if err == nil {
		// Note: this is currently not included in minikube
		err = r.Runner.Run("sudo systemctl start rktlet")
	}
	return err
}

// Disable idempotently disables rkt on a host
func (r *Rkt) Disable() error {
	err := r.Runner.Run("sudo systemctl stop rkt-api")
	if err == nil {
		err = r.Runner.Run("sudo systemctl stop rkt-metadata")
	}
	return err
}

// LoadImage loads an image into this runtime
func (r *Rkt) LoadImage(path string) error {
	glog.Infof("Loading image: %s", path)
	// Note: this command (import) does not currently exist for rkt
	return r.Runner.Run(fmt.Sprintf("sudo rkt image import %s", path))
}

// KubeletOptions returns kubelet options for a rkt
func (r *Rkt) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime":          "remote",
		"container-runtime-endpoint": fmt.Sprintf("unix://%s", r.SocketPath()),
		"image-service-endpoint":     fmt.Sprintf("unix://%s", r.SocketPath()),
		"runtime-request-timeout":    "15m",
	}
}

// ListContainers returns a list of managed by this container runtime
func (r *Rkt) ListContainers(filter string) ([]string, error) {
	return listCRIContainers(r.Runner, filter)
}

// KillContainers removes containers based on ID
func (r *Rkt) KillContainers(ids []string) error {
	return killCRIContainers(r.Runner, ids)
}

// StopContainers stops containers based on ID
func (r *Rkt) StopContainers(ids []string) error {
	return stopCRIContainers(r.Runner, ids)
}

// ContainerLogCmd returns the command to retrieve the log for a container based on ID
func (r *Rkt) ContainerLogCmd(id string, len int, follow bool) string {
	return criContainerLogCmd(id, len, follow)
}
