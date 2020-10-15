/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package tests

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"k8s.io/klog/v2"
)

// MockAPI is a struct used to mock out libmachine.API
type MockAPI struct {
	FakeStore
	CreateError   bool
	RemoveError   bool
	NotExistError bool
	SaveCalled    bool
	t             *testing.T
}

// NewMockAPI returns a new MockAPI
func NewMockAPI(t *testing.T) *MockAPI {
	t.Helper()
	m := MockAPI{
		FakeStore: FakeStore{
			Hosts: make(map[string]*host.Host),
			T:     t,
		},
		t: t,
	}
	return &m
}

// Logf logs mock interactions
func (api *MockAPI) Logf(format string, args ...interface{}) {
	if api.t == nil {
		klog.Infof(format, args...)
		return
	}
	api.t.Logf(format, args...)
}

// Close closes the API.
func (api *MockAPI) Close() error {
	return nil
}

// NewHost creates a new host.Host instance.
func (api *MockAPI) NewHost(drvName string, rawDriver []byte) (*host.Host, error) {
	var driver MockDriver
	if err := json.Unmarshal(rawDriver, &driver); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling json")
	}

	h := &host.Host{
		DriverName: drvName,
		RawDriver:  rawDriver,
		Driver:     &MockDriver{},
		Name:       fmt.Sprintf("mock-machine-%.8f", rand.Float64()),
		HostOptions: &host.Options{
			AuthOptions:  &auth.Options{},
			SwarmOptions: &swarm.Options{},
		},
	}

	api.Logf("MockAPI.NewHost: Setting profile=%q", h.Name)
	viper.Set("profile", h.Name)

	api.Logf("MockAPI.NewHost: %+v", h)
	return h, nil
}

// Load a created mock
func (api *MockAPI) Load(name string) (*host.Host, error) {
	h, err := api.FakeStore.Load(name)
	api.Logf("MockAPI.Load: %+v - %v", h, err)
	return h, err
}

// Create creates the actual host.
func (api *MockAPI) Create(h *host.Host) error {
	api.Logf("MockAPI.Create: %+v", h)
	if api.CreateError {
		return errors.New("error creating host")
	}
	// Propagate test info messages
	drv, ok := h.Driver.(*MockDriver)
	if ok {
		drv.T = api.t
	}
	if api.NotExistError {
		// initialize api.NotExistError
		api.NotExistError = false
		// reproduce ErrMachineNotExist
		drv.NotExistError = true
	}
	return h.Driver.Create()
}

// List the existing hosts.
func (api *MockAPI) List() ([]string, error) {
	return []string{}, nil
}

// Remove a host.
func (api *MockAPI) Remove(name string) error {
	api.Logf("MockAPI.Delete: %s", name)
	if api.RemoveError {
		return fmt.Errorf("error removing %s", name)
	}

	delete(api.Hosts, name)
	return nil
}

// Save saves a host to disk.
func (api *MockAPI) Save(host *host.Host) error {
	api.SaveCalled = true
	api.Logf("MockAPI.Save: %+v", host)
	return api.FakeStore.Save(host)
}

// GetMachinesDir returns the directory to store machines in.
func (api MockAPI) GetMachinesDir() string {
	return ""
}
