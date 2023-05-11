/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package persisttest

import (
	"errors"
	"testing"

	"k8s.io/minikube/pkg/libmachine/libmachine/host"
)

type FakeStore struct {
	Hosts                                           map[string]*host.Host
	ExistsErr, ListErr, LoadErr, RemoveErr, SaveErr error
	T                                               *testing.T
}

// x7NOTE: changed implementation here --- [!!] fakeStore
// if you don't remember.. it means it's ok to remove this commnt
func (fs *FakeStore) Exists(name string) (bool, error) {
	if fs.ExistsErr != nil {
		return false, fs.ExistsErr
	}

	_, ok := fs.Hosts[name]
	return ok, nil
}

func (fs *FakeStore) List() ([]string, error) {
	names := []string{}
	for _, h := range fs.Hosts {
		names = append(names, h.Name)
	}
	return names, fs.ListErr
}

func (fs *FakeStore) Load(name string) (*host.Host, error) {
	if fs.LoadErr != nil {
		return nil, fs.LoadErr
	}

	if hst, ok := fs.Hosts[name]; ok {
		return hst, nil
	}

	return nil, errors.New("no host was found inside the fakestore")
}

func (fs *FakeStore) Remove(name string) error {
	if fs.RemoveErr != nil {
		return fs.RemoveErr
	}

	delete(fs.Hosts, name)

	return nil
}

func (fs *FakeStore) Save(hst *host.Host) error {
	if fs.SaveErr == nil {
		if fs.Hosts == nil {
			fs.Hosts = make(map[string]*host.Host, 0)
		}

		fs.Hosts[hst.Name] = hst
		return nil
	}
	return fs.SaveErr
}
