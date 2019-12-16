/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"io/ioutil"
	"path/filepath"

	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// Machine contains information about a machine
type Machine struct {
	*host.Host
}

// IsValid checks if the machine has the essential info needed for a machine
func (h *Machine) IsValid() bool {
	if h == nil {
		return false
	}

	if h.Host == nil {
		return false
	}

	if h.Host.Name == "" {
		return false
	}

	if h.Host.Driver == nil {
		return false
	}

	if h.Host.HostOptions == nil {
		return false
	}

	if h.Host.RawDriver == nil {
		return false
	}
	return true
}

// List return all valid and invalid machines
// If a machine is valid or invalid is determined by the cluster.IsValid function
func List(miniHome ...string) (validMachines []*Machine, inValidMachines []*Machine, err error) {
	pDirs, err := machineDirs(miniHome...)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range pDirs {
		p, err := Load(n)
		if err != nil {
			glog.Infof("%s not valid: %v", n, err)
			inValidMachines = append(inValidMachines, p)
			continue
		}
		if !p.IsValid() {
			inValidMachines = append(inValidMachines, p)
			continue
		}
		validMachines = append(validMachines, p)
	}
	return validMachines, inValidMachines, nil
}

// Load loads a machine or throws an error if the machine could not be loadedG
func Load(name string) (*Machine, error) {
	api, err := NewAPIClient()
	if err != nil {
		return nil, err
	}

	h, err := cluster.CheckIfHostExistsAndLoad(api, name)
	if err != nil {
		return nil, err
	}

	var mm Machine
	if h != nil {
		mm.Host = h
	} else {
		return nil, errors.New("host is nil")
	}

	return &mm, nil
}

func machineDirs(miniHome ...string) (dirs []string, err error) {
	miniPath := localpath.MiniPath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}
	mRootDir := filepath.Join(miniPath, "machines")
	items, err := ioutil.ReadDir(mRootDir)
	for _, f := range items {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}
	return dirs, err
}
