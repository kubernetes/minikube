/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package libmachinetest

import (
	"k8s.io/minikube/pkg/libmachine"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/host"
	"k8s.io/minikube/pkg/libmachine/mcnerror"
	"k8s.io/minikube/pkg/libmachine/state"
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
