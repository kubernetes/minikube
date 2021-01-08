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
	"fmt"
	"sync"

	"github.com/docker/machine/libmachine/drivers"

	"k8s.io/minikube/pkg/minikube/config"
)

// Priority is how we determine what driver to default to
type Priority int

const (
	// Unknown is when there is no status check available
	Unknown Priority = iota
	// Obsolete is when a driver has been removed
	Obsolete
	// Unhealthy is when a driver does not pass health checks
	Unhealthy
	// Experimental is when a driver is not officially supported because it's still experimental
	Experimental
	// Discouraged is when a driver has caveats that preclude it's recommendation
	Discouraged
	// Deprecated is when a driver has been formally deprecated
	Deprecated
	// Fallback is when a driver works well, but may not be high performance
	Fallback
	// Default is what most 3rd party drivers are
	Default
	// Preferred is for drivers that use a native hypervisor interface
	Preferred
	// HighlyPreferred is the ultimate driver preferences
	HighlyPreferred
)

// Registry contains all the supported driver definitions on the host
type Registry interface {
	// Register a driver in registry
	Register(driver DriverDef) error

	// Driver returns the registered driver from a given name
	Driver(name string) (DriverDef, error)

	// List
	List() []DriverDef
}

// Configurator emits a struct to be marshalled into JSON for Machine Driver
type Configurator func(config.ClusterConfig, config.Node) (interface{}, error)

// Loader is a function that loads a byte stream and creates a driver.
type Loader func() drivers.Driver

// StatusChecker checks if a driver is available, offering a
type StatusChecker func() State

// State is the current state of the driver and its dependencies
type State struct {
	Installed        bool
	Healthy          bool
	Running          bool // it at least appears to be running
	NeedsImprovement bool // healthy but could be improved
	Error            error
	Fix              string
	Doc              string
}

// DriverDef defines how to initialize and load a machine driver
type DriverDef struct {
	// Name of the machine driver. It has to be unique.
	Name string

	// Alias contains a list of machine driver aliases. Each alias should also be unique.
	Alias []string

	// Config is a function that emits a configured driver struct
	Config Configurator

	// Init is a function that initializes a machine driver, if built-in to the minikube binary
	Init Loader

	// Status returns the installation status of the driver
	Status StatusChecker

	// Priority returns the prioritization for selecting a driver by default.
	Priority Priority
}

// Empty returns true if the driver is nil
func (d DriverDef) Empty() bool {
	return d.Name == ""
}

func (d DriverDef) String() string {
	return d.Name
}

type driverRegistry struct {
	drivers        map[string]DriverDef
	driversByAlias map[string]DriverDef
	lock           sync.RWMutex
}

func newRegistry() *driverRegistry {
	return &driverRegistry{
		drivers:        make(map[string]DriverDef),
		driversByAlias: make(map[string]DriverDef),
	}
}

// Register registers a driver
func (r *driverRegistry) Register(def DriverDef) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.drivers[def.Name]; ok {
		return fmt.Errorf("%q is already registered: %+v", def.Name, def)
	}

	r.drivers[def.Name] = def

	for _, alias := range def.Alias {
		if _, ok := r.driversByAlias[alias]; ok {
			return fmt.Errorf("alias %q is already registered: %+v", alias, def)
		}
		r.driversByAlias[alias] = def
	}
	return nil
}

// List returns a list of registered drivers
func (r *driverRegistry) List() []DriverDef {
	r.lock.RLock()
	defer r.lock.RUnlock()

	result := make([]DriverDef, 0, len(r.drivers))

	for _, def := range r.drivers {
		result = append(result, def)
	}

	return result
}

// Driver returns a driver given a name
func (r *driverRegistry) Driver(name string) DriverDef {
	r.lock.RLock()
	defer r.lock.RUnlock()

	def, ok := r.drivers[name]
	if ok {
		return def
	}

	// Check if we have driver def with name as alias
	return r.driversByAlias[name]
}
