/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package docker

import (
	"encoding/json"
	"os/exec"
	"path"

	"github.com/opencontainers/go-digest"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
)

const (
	referenceStorePath = "/var/lib/docker/image/overlay2/repositories.json"
)

// Storage keeps track of reference stores
type Storage struct {
	refStores []ReferenceStore
	runner    command.Runner
}

// ReferenceStore stores references to images in repositories.json
// used by the docker daemon to name images
// taken from "github.com/docker/docker/reference/store.go"
type ReferenceStore struct {
	Repositories map[string]repository
}

type repository map[string]digest.Digest

// NewStorage returns a new storage type
func NewStorage(runner command.Runner) *Storage {
	return &Storage{
		runner: runner,
	}
}

// Save saves the current reference store in memory
func (s *Storage) Save() error {
	// get the contents of repositories.json in minikube
	// if this command fails, assume the file doesn't exist
	rr, err := s.runner.RunCmd(exec.Command("sudo", "cat", referenceStorePath))
	if err != nil {
		klog.Infof("repositories.json doesn't exist: %v", err)
		return nil
	}
	contents := rr.Stdout.Bytes()
	var rs ReferenceStore
	if err := json.Unmarshal(contents, &rs); err != nil {
		return err
	}
	s.refStores = append(s.refStores, rs)
	return nil
}

// Update merges all reference stores and updates repositories.json
func (s *Storage) Update() error {
	// in case we didn't overwrite respoitories.json, do nothing
	if len(s.refStores) == 1 {
		return nil
	}
	// merge reference stores
	merged := s.mergeReferenceStores()

	// write to file in minikube
	contents, err := json.Marshal(merged)
	if err != nil {
		return err
	}

	asset := assets.NewMemoryAsset(contents, path.Dir(referenceStorePath), path.Base(referenceStorePath), "0644")
	return s.runner.Copy(asset)
}

func (s *Storage) mergeReferenceStores() ReferenceStore {
	merged := ReferenceStore{
		Repositories: map[string]repository{},
	}
	// otherwise, merge together reference stores
	for _, rs := range s.refStores {
		for k, v := range rs.Repositories {
			merged.Repositories[k] = v
		}
	}
	return merged
}
