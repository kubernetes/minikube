/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package sysinit

import (
	"os/exec"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
)

var cachedSystemdCheck *bool

// Runner is the subset of command.Runner this package consumes
type Runner interface {
	RunCmd(cmd *exec.Cmd) (*command.RunResult, error)
}

// Manager is a common interface for init systems
type Manager interface {
	// Name returns the name of the init manager
	Name() string

	// Active returns if a service is active
	Active(string) bool

	// Disable disables a service
	Disable(string) error

	// Disable disables a service and stops it right after.
	DisableNow(string) error

	// Mask prevents a service from being started
	Mask(string) error

	// Enable enables a service
	Enable(string) error

	// EnableNow enables a service and starts it right after.
	EnableNow(string) error

	// Unmask allows a service to be started
	Unmask(string) error

	// Start starts a service idempotently
	Start(string) error

	// Restart restarts a service
	Restart(string) error

	// Reload restarts a service
	Reload(string) error

	// Stop stops a service
	Stop(string) error

	// ForceStop stops a service with prejudice
	ForceStop(string) error

	// GenerateInitShim generates any additional init files required for this service
	GenerateInitShim(svc string, binary string, unit string) ([]assets.CopyableFile, error)
}

// New returns an appropriately configured service manager
func New(r Runner) Manager {
	// If we are not provided a runner, we can't do anything anyways
	if r == nil {
		return nil
	}

	var systemd bool

	// Caching the result is important, as this manager may be created in many places,
	// and ssh calls are expensive on some drivers, such as Docker.
	if cachedSystemdCheck != nil {
		systemd = *cachedSystemdCheck
	} else {
		systemd = usesSystemd(r)
		cachedSystemdCheck = &systemd
	}

	if systemd {
		return &Systemd{r: r}
	}
	return &OpenRC{r: r}
}
