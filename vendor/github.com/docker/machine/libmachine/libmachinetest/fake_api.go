package libmachinetest

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/state"
)

type FakeAPI struct {
	Hosts []*host.Host
}

func (api *FakeAPI) NewPluginDriver(string, []byte) (drivers.Driver, error) {
	return nil, nil
}

func (api *FakeAPI) Close() error {
	return nil
}

func (api *FakeAPI) NewHost(driverName string, rawDriver []byte) (*host.Host, error) {
	return nil, nil
}

func (api *FakeAPI) Create(h *host.Host) error {
	return nil
}

func (api *FakeAPI) Exists(name string) (bool, error) {
	for _, host := range api.Hosts {
		if name == host.Name {
			return true, nil
		}
	}

	return false, nil
}

func (api *FakeAPI) List() ([]string, error) {
	return []string{}, nil
}

func (api *FakeAPI) Load(name string) (*host.Host, error) {
	for _, host := range api.Hosts {
		if name == host.Name {
			return host, nil
		}
	}

	return nil, mcnerror.ErrHostDoesNotExist{
		Name: name,
	}
}

func (api *FakeAPI) Remove(name string) error {
	newHosts := []*host.Host{}

	for _, host := range api.Hosts {
		if name != host.Name {
			newHosts = append(newHosts, host)
		}
	}

	api.Hosts = newHosts

	return nil
}

func (api *FakeAPI) Save(host *host.Host) error {
	return nil
}

func (api FakeAPI) GetMachinesDir() string {
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
