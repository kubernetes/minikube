/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package drivers

import (
	"errors"
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
)

// Driver defines how a host is created and controlled. Different types of
// driver represent different ways hosts can be created (e.g. different
// hypervisors, different cloud providers)
type Driver interface {
	// GetRunner gets the adequate implementation of the command runner, based
	// on the driver (e.g. ssh runner, docker exec, ...)
	GetRunner() (runner.Runner, error)

	// RunCmd is a shortcut to run a command inside the linux machine
	RunCmd(*exec.Cmd) (*runner.RunResult, error)

	// IsManaged once the linux machine is created,
	// is minikube responsible for the provisioning?
	// otherwise we assume that the machine has already all the components installed.
	IsManaged() bool

	// IsContainerBased returns true if this driver creates a container based machine
	IsContainerBased() bool

	// IsISOBased returns true if this driver creates an ISO based machine
	IsISOBased() bool

	// Create a host using the driver's config
	CreateMachine() error

	// DriverName returns the name of the driver
	DriverName() string

	// GetCreateFlags returns the mcnflag.Flag slice representing the flags
	// that can be set, their descriptions and defaults.
	GetCreateFlags() []mcnflag.Flag

	// GetIP returns an IP or hostname that this host is available at
	// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
	GetIP() (string, error)

	// GetMachineName returns the name of the machine
	GetMachineName() string

	// GetSSHHostname returns hostname for use with ssh
	// required since we're using this for ssl/tls authentication setup during provisioning
	GetSSHHostname() (string, error)

	// GetSSHKeyPath returns key path for use with ssh
	GetSSHKeyPath() string

	// GetSSHPort returns port for use with ssh
	GetSSHPort() (int, error)

	// GetSSHUsername returns username for use with ssh
	GetSSHUsername() string

	// GetURL returns a Docker compatible host URL for connecting to this host
	// e.g. tcp://1.2.3.4:2376
	GetURL() (string, error)

	// GetMachineState returns the state that the linux machine is in (running, stopped, etc)
	GetMachineState() (state.State, error)

	// KillMachine stops the linux machine forcefully
	KillMachine() error

	// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
	PreCreateCheck() error

	// RemoveMachine removes the linux machine
	RemoveMachine() error

	// RestartMachine may just call Stop(); Start() if the provider does not
	// have any special restart behaviour.
	RestartMachine() error

	// SetConfigFromFlags configures the driver with the object that was returned
	// by RegisterCreateFlags
	SetConfigFromFlags(opts DriverOptions) error

	// StartMachine starts the linux machine we're running kubernetes into
	StartMachine() error

	// StopMachine gracefully stops the linux machine
	StopMachine() error
}

var ErrHostIsNotRunning = errors.New("host is not running")

type DriverOptions interface {
	String(key string) string
	StringSlice(key string) []string
	Int(key string) int
	Bool(key string) bool
}

func MachineInState(d Driver, desiredState state.State) func() bool {
	return func() bool {
		currentState, err := d.GetMachineState()
		if err != nil {
			log.Debugf("Error getting machine state: %s", err)
		}
		if currentState == desiredState {
			return true
		}
		return false
	}
}

// MustBeRunning will return an error if the machine is not in a running state.
func MustBeRunning(d Driver) error {
	s, err := d.GetMachineState()
	if err != nil {
		return err
	}

	if s != state.Running {
		return ErrHostIsNotRunning
	}

	return nil
}
