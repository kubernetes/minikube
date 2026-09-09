/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package registry

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"sync"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/minikube/translate"
)

const (
	// Podman is Kubernetes in container using podman driver
	Podman = "podman"
	// Docker is Kubernetes in container using docker driver
	Docker = "docker"
	// Mock driver
	Mock = "mock"
	// None driver
	None = "none"
)

// IsKIC checks if the driver is a Kubernetes in container
func IsKIC(name string) bool {
	return name == Docker || name == Podman
}

// IsMock checks if the driver is a mock
func IsMock(name string) bool {
	return name == Mock
}

// IsVM checks if the driver is a VM
func IsVM(name string) bool {
	if IsKIC(name) || IsMock(name) || BareMetal(name) {
		return false
	}
	return true
}

// BareMetal returns if this driver is unisolated
func BareMetal(name string) bool {
	return name == None || name == Mock
}

var (
	// globalRegistry is a globally accessible driver registry
	globalRegistry = newRegistry()
)

// DriverState is metadata relating to a driver and status
type DriverState struct {
	// Name is the name of the driver used internally
	Name string
	// Default drivers are selected automatically
	Default bool
	// Preference is the original priority from driver
	Preference Priority
	// Priority is the effective priority with health
	Priority Priority
	// State is the state of driver and dependencies
	State State
	// Rejection is why we chose not to use this driver
	Rejection string
	// Suggestion is how the user could improve health
	Suggestion string
}

func (d DriverState) String() string {
	if d.Priority == Experimental {
		experimental := translate.T("experimental")
		return fmt.Sprintf("%s (%s)", d.Name, experimental)
	}
	return d.Name
}

// List lists drivers in global registry
func List() []DriverDef {
	return globalRegistry.List()
}

// Register registers driver with the global registry
func Register(driver DriverDef) error {
	return globalRegistry.Register(driver)
}

// Driver gets a named driver from the global registry
func Driver(name string) DriverDef {
	return globalRegistry.Driver(name)
}

// Available returns a list of available drivers in the global registry
func Available(vm bool, options *run.CommandOptions) []DriverState {
	klog.Infof("Querying for installed drivers using PATH=%s", os.Getenv("PATH"))

	type probeResult struct {
		driver DriverDef
		state  State
	}

	drivers := globalRegistry.List()
	results := make(chan probeResult, len(drivers))
	var wg sync.WaitGroup

	for _, d := range drivers {
		if vm && !IsVM(d.Name) {
			continue
		}
		if d.Status == nil {
			klog.Errorf("%q does not implement Status", d.Name)
			continue
		}
		wg.Add(1)
		go func(d DriverDef) {
			defer wg.Done()
			results <- probeResult{driver: d, state: probe(d, options)}
		}(d)
	}

	wg.Wait()
	close(results)

	sts := make([]DriverState, 0, len(drivers))
	for r := range results {
		s := r.state
		d := r.driver
		klog.Infof("%s default: %v priority: %d, state: %+v", d.Name, d.Default, d.Priority, s)

		preference := d.Priority
		priority := d.Priority
		if !s.Healthy {
			priority = Unhealthy
		}
		sts = append(sts, DriverState{Name: d.Name, Default: d.Default, Preference: preference, Priority: priority, State: s})
	}

	// Descending priority for predictability
	slices.SortFunc(sts, func(a, b DriverState) int {
		return cmp.Compare(b.Priority, a.Priority)
	})
	return sts
}

// probe runs a driver's Status check with an independent timeout so Available
// can query every driver at once. Wall-clock time is therefore about the
// slowest probe, not the sum of all probes.
func probe(d DriverDef, options *run.CommandOptions) State {
	timeout := d.ProbeTimeout

	stateChannel := make(chan State, 1)
	go func() {
		stateChannel <- d.Status(options)
	}()

	select {
	case <-time.After(timeout):
		klog.Infof("%s status check timeout!", d.Name)
		return State{}
	case s := <-stateChannel:
		return s
	}
}

// Status returns the state of a driver within the global registry
func Status(name string, options *run.CommandOptions) State {
	d := globalRegistry.Driver(name)
	if d.Empty() {
		return State{}
	}
	return d.Status(options)
}
