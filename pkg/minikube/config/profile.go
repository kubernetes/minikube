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

package config

import (
	"io/ioutil"
	"path/filepath"

	"k8s.io/minikube/pkg/minikube/constants"
)

// isValid checks if the profile has the essential info needed for a profile
func (p *Profile) isValid() bool {
	if p.Config.MachineConfig.VMDriver == "" {
		return false
	}
	if p.Config.KubernetesConfig.KubernetesVersion == "" {
		return false
	}
	return true
}

// ListProfiles returns all valid and invalid (if any) minikube profiles
// invalidPs are the profiles that have a directory or config file but not usable
// invalidPs would be suggeted to be deleted
func ListProfiles(miniHome ...string) (validPs []*Profile, inValidPs []*Profile, err error) {
	pDirs, err := profileDirs(miniHome...)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range pDirs {
		p, err := loadProfile(n, miniHome...)
		if err != nil {
			inValidPs = append(inValidPs, p)
			continue
		}
		if !p.isValid() {
			inValidPs = append(inValidPs, p)
			continue
		}
		validPs = append(validPs, p)
	}
	return validPs, inValidPs, nil
}

// loadProfile loads type Profile based on its name
func loadProfile(name string, miniHome ...string) (*Profile, error) {
	cfg, err := DefaultLoader.LoadConfigFromFile(name, miniHome...)
	p := &Profile{
		Name:   name,
		Config: cfg,
	}
	return p, err
}

// profileDirs gets all the folders in the user's profiles folder regardless of valid or invalid config
func profileDirs(miniHome ...string) (dirs []string, err error) {
	miniPath := constants.GetMinipath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}
	pRootDir := filepath.Join(miniPath, "profiles")
	items, err := ioutil.ReadDir(pRootDir)
	for _, f := range items {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}
	return dirs, err
}
