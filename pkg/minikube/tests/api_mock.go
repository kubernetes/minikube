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

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
)

// MockAPI is a struct used to mock out libmachine.API
type MockAPI struct {
	FakeStore
	CreateError bool
	RemoveError bool
	SaveCalled  bool
}

// NewMockAPI returns a new MockAPI
func NewMockAPI() *MockAPI {
	m := MockAPI{
		FakeStore: FakeStore{
			Hosts: make(map[string]*host.Host),
		},
	}
	return &m
}

// Close closes the API.
func (api *MockAPI) Close() error {
	return nil
}

// NewHost creates a new host.Host instance.
func (api *MockAPI) NewHost(driverName string, rawDriver []byte) (*host.Host, error) {
	var driver MockDriver
	if err := json.Unmarshal(rawDriver, &driver); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling json")
	}
	h := &host.Host{
		DriverName:  driverName,
		RawDriver:   rawDriver,
		Driver:      &MockDriver{},
		Name:        driver.GetMachineName(),
		HostOptions: &host.Options{AuthOptions: &auth.Options{}},
	}
	return h, nil
}

// Create creates the actual host.
func (api *MockAPI) Create(h *host.Host) error {
	if api.CreateError {
		return errors.New("error creating host")
	}
	return h.Driver.Create()
}

// List the existing hosts.
func (api *MockAPI) List() ([]string, error) {
	return []string{}, nil
}

// Remove a host.
func (api *MockAPI) Remove(name string) error {
	if api.RemoveError {
		return fmt.Errorf("error removing %s", name)
	}

	delete(api.Hosts, name)
	return nil
}

// Save saves a host to disk.
func (api *MockAPI) Save(host *host.Host) error {
	api.SaveCalled = true
	return api.FakeStore.Save(host)
}

// GetMachinesDir returns the directory to store machines in.
func (api MockAPI) GetMachinesDir() string {
	return ""
}
