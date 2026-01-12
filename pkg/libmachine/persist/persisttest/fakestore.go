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

package persisttest

import "k8s.io/minikube/pkg/libmachine/host"

type FakeStore struct {
	Hosts                                           []*host.Host
	ExistsErr, ListErr, LoadErr, RemoveErr, SaveErr error
}

func (fs *FakeStore) Exists(name string) (bool, error) {
	if fs.ExistsErr != nil {
		return false, fs.ExistsErr
	}
	for _, h := range fs.Hosts {
		if h.Name == name {
			return true, nil
		}
	}

	return false, nil
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
	for _, h := range fs.Hosts {
		if h.Name == name {
			return h, nil
		}
	}

	return nil, nil
}

func (fs *FakeStore) Remove(name string) error {
	if fs.RemoveErr != nil {
		return fs.RemoveErr
	}
	for i, h := range fs.Hosts {
		if h.Name == name {
			fs.Hosts = append(fs.Hosts[:i], fs.Hosts[i+1:]...)
			return nil
		}
	}
	return nil
}

func (fs *FakeStore) Save(host *host.Host) error {
	if fs.SaveErr == nil {
		fs.Hosts = append(fs.Hosts, host)
		return nil
	}
	return fs.SaveErr
}
