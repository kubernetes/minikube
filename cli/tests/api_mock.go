package tests

import (
	"encoding/json"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/state"
)

// Close closes the API.
func (api *MockAPI) Close() error {
	return nil
}

// NewHost creates a new host.Host instance.
func (api *MockAPI) NewHost(driverName string, rawDriver []byte) (*host.Host, error) {
	var driver MockDriver
	if err := json.Unmarshal(rawDriver, &driver); err != nil {
		return nil, err
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
	return h.Driver.Create()
}

// Exists determines if the host already exists.
func (api *MockAPI) Exists(name string) (bool, error) {
	for _, host := range api.Hosts {
		if name == host.Name {
			return true, nil
		}
	}

	return false, nil
}

// List the existing hosts.
func (api *MockAPI) List() ([]string, error) {
	return []string{}, nil
}

// Load loads a host from disk.
func (api *MockAPI) Load(name string) (*host.Host, error) {
	for _, host := range api.Hosts {
		if name == host.Name {
			return host, nil
		}
	}

	return nil, mcnerror.ErrHostDoesNotExist{
		Name: name,
	}
}

// Remove a host.
func (api *MockAPI) Remove(name string) error {
	newHosts := []*host.Host{}

	for _, host := range api.Hosts {
		if name != host.Name {
			newHosts = append(newHosts, host)
		}
	}

	api.Hosts = newHosts

	return nil
}

// Save saves a host to disk.
func (api *MockAPI) Save(host *host.Host) error {
	api.Hosts = append(api.Hosts, host)
	return nil
}

// GetMachinesDir returns the directory to store machines in.
func (api MockAPI) GetMachinesDir() string {
	return ""
}

// State returns the state of a host.
func State(api libmachine.API, name string) state.State {
	host, _ := api.Load(name)
	machineState, _ := host.Driver.GetState()
	return machineState
}

// Exists tells whether a named host exists.
func Exists(api libmachine.API, name string) bool {
	exists, _ := api.Exists(name)
	return exists
}
