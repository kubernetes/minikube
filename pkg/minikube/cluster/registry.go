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

package cluster

import (
	"sync"

	"github.com/pkg/errors"
)

var (
	// ErrDriverNameExist is the error returned when trying to register a driver
	// which already exists in registry
	ErrDriverNameExist = errors.New("registry: duplicated driver name")

	// ErrDriverNotFound is the error returned when driver of a given name does
	// not exist in registry
	ErrDriverNotFound = errors.New("registry: driver not found")
)

// DriverFunc creates a Driver from MachineConfig
type DriverFunc func(MachineConfig) RawDriver

// RawDriver is the configuration of a driver that you can marshal into json bytes
// and pass to libmachine's NewHost method
type RawDriver interface{}

// Registry contains all the supported driver types on the host
type Registry interface {
	// Register a driver in registry
	Register(name string, f DriverFunc) error

	// Driver returns the registered driver from a given name
	Driver(name string) (DriverFunc, error)

	DriverLister
}

// DriverLister lists the name of the supported drivers
type DriverLister interface {
	// List lists all the driver types
	List() []string
}

type driverRegistry struct {
	drivers map[string]DriverFunc
	lock    sync.Mutex
}

func createRegistry() *driverRegistry {
	return &driverRegistry{
		drivers: make(map[string]DriverFunc),
	}
}

var (
	registry = createRegistry()
)

func ListDrivers() []string {
	return registry.List()
}

func (r *driverRegistry) Register(name string, f DriverFunc) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.drivers[name]; ok {
		return ErrDriverNameExist
	}

	r.drivers[name] = f
	return nil
}

func (r *driverRegistry) List() []string {
	r.lock.Lock()
	defer r.lock.Unlock()

	result := make([]string, 0, len(r.drivers))

	for name := range r.drivers {
		result = append(result, name)
	}

	return result
}

func (r *driverRegistry) Driver(name string) (DriverFunc, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if driver, ok := r.drivers[name]; ok {
		return driver, nil
	}

	return nil, ErrDriverNotFound
}
