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
	"os"
	"sort"

	"github.com/golang/glog"
)

var (
	// globalRegistry is a globally accessible driver registry
	globalRegistry = newRegistry()
)

// DriverState is metadata relating to a driver and status
type DriverState struct {
	Name     string
	Priority Priority
	State    State
}

func (d DriverState) String() string {
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
func Available() []DriverState {
	sts := []DriverState{}
	glog.Infof("Querying for installed drivers using PATH=%s", os.Getenv("PATH"))

	for _, d := range globalRegistry.List() {
		if d.Status == nil {
			glog.Errorf("%q does not implement Status", d.Name)
			continue
		}
		s := d.Status()
		glog.Infof("%s priority: %d, state: %+v", d.Name, d.Priority, s)

		priority := d.Priority
		if !s.Healthy {
			priority = Unhealthy
		}

		sts = append(sts, DriverState{Name: d.Name, Priority: priority, State: s})
	}

	// Descending priority for predictability
	sort.Slice(sts, func(i, j int) bool {
		return sts[i].Priority > sts[j].Priority
	})
	return sts
}

// Status returns the state of a driver within the global registry
func Status(name string) State {
	d := globalRegistry.Driver(name)
	if d.Empty() {
		return State{}
	}
	return d.Status()
}
