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
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
)

var (
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

// ConfigFactory is a function that creates a driver config from MachineConfig
type ConfigFactory func(config.MachineConfig) interface{}

// DriverFactory is a function that load a byte stream and create a driver
type DriverFactory func() drivers.Driver

// DriverDef defines a machine driver metadata. It tells minikube how to initialize
// and load drivers.
type DriverDef struct {
	// Name of the machine driver. It has to be unique.
	Name string

	// BuiltIn indicates if the driver is builtin minikube binary, or the driver is
	// triggered through RPC.
	Builtin bool

	// ConfigCreator generate a raw driver object by minikube's machine config.
	ConfigCreator ConfigFactory

	// DriverCreator is the factory method that creates a machine driver instance.
	DriverCreator DriverFactory
}

func (d DriverDef) String() string {
	return fmt.Sprintf("{name: %s, builtin: %t}", d.Name, d.Builtin)
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

var (
	registry = createRegistry()
)

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

func (r *driverRegistry) Register(def DriverDef) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.drivers[def.Name]; ok {
		return ErrDriverNameExist
	}

	r.drivers[def.Name] = def
	return nil
}

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
