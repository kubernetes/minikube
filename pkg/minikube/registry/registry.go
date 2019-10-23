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
	"sync"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/minikube/config"
)

type Priority int

const (
	// Priority for default driver selection
	Unknown Priority = iota
	Discouraged
	Deprecated
	Fallback
	Default
	Preferred
	StronglyPreferred
)

var (
	registry = createRegistry()

	// ErrDriverNameExist is the error returned when trying to register a driver
	// which already exists in registry
	ErrDriverNameExist = errors.New("registry: duplicated driver name")

	// ErrDriverNotFound is the error returned when driver of a given name does
	// not exist in registry
	ErrDriverNotFound = errors.New("registry: driver not found")
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
type Configurator func(config.MachineConfig) interface{}

// Loader is a function that loads a byte stream and creates a driver.
type Loader func() drivers.Driver

// Status checks if a driver is available, offering a
type StatusChecker func() State

// State is the current state of the driver and its dependencies
type State struct {
	Installed bool
	Healthy   bool
	Error     error
	Fix       string
	Doc       string
}

// State is metadata relating to a driver and status
type DriverState struct {
	Name     string
	Priority Priority
	State    State
}

func (d DriverState) String() string {
	return d.Name
}

// DriverDef defines how to initialize and load a machine driver
type DriverDef struct {
	// Name of the machine driver. It has to be unique.
	Name string

	// Config is a function that emits a configured driver struct
	Config Configurator

	// Init is a function that initializes a machine driver, if built-in to the minikube binary
	Init Loader

	// Status returns the installation status of the driver
	Status StatusChecker

	// Priority returns the prioritization for selecting a driver by default.
	Priority Priority
}

func (d DriverDef) String() string {
	return d.Name
}

type driverRegistry struct {
	drivers map[string]DriverDef
	lock    sync.Mutex
}

func createRegistry() *driverRegistry {
	return &driverRegistry{
		drivers: make(map[string]DriverDef),
	}
}

// ListDrivers lists all drivers in registry
func ListDrivers() []DriverDef {
	return registry.List()
}

// Register registers driver
func Register(driver DriverDef) error {
	return registry.Register(driver)
}

// Driver gets a named driver
func Driver(name string) (DriverDef, error) {
	return registry.Driver(name)
}

// Installed returns a list of installed drivers
func Installed() []DriverState {
	sts := []DriverState{}
	for _, d := range registry.List() {
		if d.Status == nil {
			glog.Errorf("%q does not implement Status", d.Name)
			continue
		}
		s := d.Status()
		if s.Installed {
			sts = append(sts, DriverState{Name: d.Name, Priority: d.Priority, State: s})
		}
	}
	return sts
}

// Status returns the state of a driver
func Status(name string) (State, error) {
	d, err := registry.Driver(name)
	if err != nil {
		return State{}, err
	}
	return d.Status(), nil
}

// Register registers a driver with minikube
func (r *driverRegistry) Register(def DriverDef) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.drivers[def.Name]; ok {
		return ErrDriverNameExist
	}

	r.drivers[def.Name] = def
	return nil
}

// List returns a list of registered drivers
func (r *driverRegistry) List() []DriverDef {
	r.lock.Lock()
	defer r.lock.Unlock()

	result := make([]DriverDef, 0, len(r.drivers))

	for _, def := range r.drivers {
		result = append(result, def)
	}

	return result
}

func (r *driverRegistry) Driver(name string) (DriverDef, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if driver, ok := r.drivers[name]; ok {
		return driver, nil
	}

	return DriverDef{}, ErrDriverNotFound
}
