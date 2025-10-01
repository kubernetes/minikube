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

package persist

import (
	"k8s.io/minikube/pkg/libmachine/host"
)

type Store interface {
	// Exists returns whether a machine exists or not
	Exists(name string) (bool, error)

	// List returns a list of all hosts in the store
	List() ([]string, error)

	// Load loads a host by name
	Load(name string) (*host.Host, error)

	// Remove removes a machine from the store
	Remove(name string) error

	// Save persists a machine in the store
	Save(host *host.Host) error
}

func LoadHosts(s Store, hostNames []string) ([]*host.Host, map[string]error) {
	loadedHosts := []*host.Host{}
	errors := map[string]error{}

	for _, hostName := range hostNames {
		h, err := s.Load(hostName)
		if err != nil {
			errors[hostName] = err
		} else {
			loadedHosts = append(loadedHosts, h)
		}
	}

	return loadedHosts, errors
}

func LoadAllHosts(s Store) ([]*host.Host, map[string]error, error) {
	hostNames, err := s.List()
	if err != nil {
		return nil, nil, err
	}
	loadedHosts, hostInError := LoadHosts(s, hostNames)
	return loadedHosts, hostInError, nil
}
