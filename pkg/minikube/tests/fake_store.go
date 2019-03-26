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

package tests

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
)

// FakeStore implements persist.Store from libmachine
type FakeStore struct {
	Hosts map[string]*host.Host
}

// Exists determines if the host already exists.
func (s *FakeStore) Exists(name string) (bool, error) {
	_, ok := s.Hosts[name]
	return ok, nil
}

// List returns the list of hosts.
func (s *FakeStore) List() ([]string, error) {
	hostNames := []string{}
	for h := range s.Hosts {
		hostNames = append(hostNames, h)
	}
	return hostNames, nil
}

// Load loads a host from disk.
func (s *FakeStore) Load(name string) (*host.Host, error) {
	h, ok := s.Hosts[name]
	if !ok {
		return nil, mcnerror.ErrHostDoesNotExist{
			Name: name,
		}

	}
	return h, nil
}

// Remove removes a machine from the store
func (s *FakeStore) Remove(name string) error {
	_, ok := s.Hosts[name]
	if !ok {
		return mcnerror.ErrHostDoesNotExist{
			Name: name,
		}

	}
	delete(s.Hosts, name)
	return nil
}

// Save persists a machine in the store
func (s *FakeStore) Save(host *host.Host) error {
	s.Hosts[host.Name] = host
	return nil
}
