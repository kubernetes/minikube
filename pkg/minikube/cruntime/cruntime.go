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

// Package cruntime contains code specific to container runtimes
package cruntime

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// CommandRunner is the subset of bootstrapper.CommandRunner this package consumes
type CommandRunner interface {
	Run(string) error
	CombinedOutput(string) (string, error)
}

// Manager is a common interface for container runtimes
type Manager interface {
	// Name is a human readable name for a runtime
	Name() string
	// Version retrieves the current version of this runtime
	Version() (string, error)
	// Enable idempotently enables this runtime on a host
	Enable() error
	// Disable idempotently disables this runtime on a host
	Disable() error
	// Active returns whether or not a runtime is active on a host
	Active() bool
	// Available returns an error if it is not possible to use this runtime on a host
	Available() error

	// KubeletOptions returns kubelet options for a runtime.
	KubeletOptions() map[string]string
	// SocketPath returns the path to the socket file for a given runtime
	SocketPath() string
	// DefaultCNI returns whether to use CNI networking by default
	DefaultCNI() bool

	// Load an image idempotently into the runtime on a host
	LoadImage(string) error

	// ListContainers returns a list of managed by this container runtime
	ListContainers(string) ([]string, error)
	// KillContainers removes containers based on ID
	KillContainers([]string) error
	// StopContainers stops containers based on ID
	StopContainers([]string) error
	// ContainerLogCmd returns the command to retrieve the log for a container based on ID
	ContainerLogCmd(string, int, bool) string
}

// Config is runtime configuration
type Config struct {
	// Type of runtime to create ("docker, "crio", etc)
	Type string
	// Custom path to a socket file
	Socket string
	// Runner is the CommandRunner object to execute commands with
	Runner CommandRunner
}

// New returns an appropriately configured runtime
func New(c Config) (Manager, error) {
	switch c.Type {
	case "", "docker":
		return &Docker{Socket: c.Socket, Runner: c.Runner}, nil
	case "crio", "cri-o":
		return &CRIO{Socket: c.Socket, Runner: c.Runner}, nil
	case "containerd":
		return &Containerd{Socket: c.Socket, Runner: c.Runner}, nil
	default:
		return nil, fmt.Errorf("unknown runtime type: %q", c.Type)
	}
}

// disableOthers disables all other runtimes except for me.
func disableOthers(me Manager, cr CommandRunner) error {
	// valid values returned by manager.Name()
	runtimes := []string{"containerd", "crio", "docker"}
	for _, name := range runtimes {
		r, err := New(Config{Type: name, Runner: cr})
		if err != nil {
			return fmt.Errorf("runtime(%s): %v", name, err)
		}

		// Don't disable myself.
		if r.Name() == me.Name() {
			continue
		}
		// runtime is already disabled, nothing to do.
		if !r.Active() {
			continue
		}
		if err = r.Disable(); err != nil {
			glog.Warningf("disable failed: %v", err)
		}
		// Validate that the runtime really is offline - and that Active & Disable are properly written.
		if r.Active() {
			return fmt.Errorf("%s is still active", r.Name())
		}
	}
	return nil
}

// enableIPForwarding configures IP forwarding, which is handled normally by Docker
// Context: https://github.com/kubernetes/kubeadm/issues/1062
func enableIPForwarding(cr CommandRunner) error {
	if err := cr.Run("sudo modprobe br_netfilter"); err != nil {
		return errors.Wrap(err, "br_netfilter")
	}
	if err := cr.Run("sudo sh -c \"echo 1 > /proc/sys/net/ipv4/ip_forward\""); err != nil {
		return errors.Wrap(err, "ip_forward")
	}
	return nil
}
