package machine

import (
	"encoding/json"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

type MockAPI struct {
	Hosts []*host.Host
}

type MockDriver struct {
	drivers.BaseDriver
	CurrentState state.State
}

func (driver *MockDriver) Create() error {
	driver.CurrentState = state.Running
	return nil
}

func (driver *MockDriver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{}
}

func (driver *MockDriver) GetSSHHostname() (string, error) {
	return "", nil
}

func (driver *MockDriver) GetState() (state.State, error) {
	return driver.CurrentState, nil
}

func (driver *MockDriver) GetURL() (string, error) {
	return "", nil
}

func (driver *MockDriver) Kill() error {
	driver.CurrentState = state.Stopped
	return nil
}

func (driver *MockDriver) Remove() error {
	return nil
}

func (driver *MockDriver) Restart() error {
	driver.CurrentState = state.Running
	return nil
}

func (driver *MockDriver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	return nil
}

func (driver *MockDriver) Start() error {
	driver.CurrentState = state.Running
	return nil
}

func (driver *MockDriver) Stop() error {
	driver.CurrentState = state.Stopped
	return nil
}

// func (api *MockAPI) NewPluginDriver(string, []byte) (drivers.Driver, error) {
// 	return nil, nil
// }

func (api *MockAPI) Close() error {
	return nil
}

func (api *MockAPI) NewHost(driverName string, rawDriver []byte) (*host.Host, error) {
	var driver MockDriver
	if err := json.Unmarshal(rawDriver, &driver); err != nil {
		return nil, err
	}
	h := &host.Host{
		DriverName: driverName,
		RawDriver:  rawDriver,
		Driver:     &MockDriver{},
		Name:       driver.GetMachineName(),
	}
	return h, nil
}

func (api *MockAPI) Create(h *host.Host) error {
	return h.Driver.Create()
}

func (api *MockAPI) Exists(name string) (bool, error) {
	for _, host := range api.Hosts {
		if name == host.Name {
			return true, nil
		}
	}

	return false, nil
}

func (api *MockAPI) List() ([]string, error) {
	return []string{}, nil
}

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

func (api *MockAPI) Save(host *host.Host) error {
	api.Hosts = append(api.Hosts, host)
	return nil
}

func (api MockAPI) GetMachinesDir() string {
	return ""
}

func State(api libmachine.API, name string) state.State {
	host, _ := api.Load(name)
	machineState, _ := host.Driver.GetState()
	return machineState
}

func Exists(api libmachine.API, name string) bool {
	exists, _ := api.Exists(name)
	return exists
}
