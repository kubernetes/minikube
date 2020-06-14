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

package v152

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/minikube/pkg/minikube/localpath"
)

// IsValid checks if the profile has the essential info needed for a profile
func (p *Profile) IsValid() bool {
	if p.Config == nil {
		return false
	}
	if p.Config.MachineConfig.VMDriver == "" {
		return false
	}
	if p.Config.KubernetesConfig.KubernetesVersion == "" {
		return false
	}
	return true
}

// ProfileExists returns true if there is a profile config (regardless of being valid)
func ProfileExists(name string, miniHome ...string) bool {
	miniPath := localpath.MiniPath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}

	p := profileFilePath(name, miniPath)
	_, err := os.Stat(p)
	return err == nil
}

// CreateProfile creates an profile out of the cfg and stores in $MINIKUBE_HOME/profiles/<profilename>/config.json

// ListProfiles returns all valid and invalid (if any) minikube profiles
// invalidPs are the profiles that have a directory or config file but not usable
// invalidPs would be suggested to be deleted
func ListProfiles(miniHome ...string) (validPs []*Profile, inValidPs []*Profile, err error) {
	pDirs, err := profileDirs(miniHome...)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range pDirs {
		p, err := LoadProfile(n, miniHome...)
		if err != nil {
			inValidPs = append(inValidPs, p)
			continue
		}
		if !p.IsValid() {
			inValidPs = append(inValidPs, p)
			continue
		}
		validPs = append(validPs, p)
	}
	return validPs, inValidPs, nil
}

// LoadProfile loads type Profile based on its name
func LoadProfile(name string, miniHome ...string) (*Profile, error) {
	cfg, err := DefaultLoader.LoadConfigFromFile(name, miniHome...)
	p := &Profile{
		Name:   name,
		Config: cfg,
	}
	return p, err
}

// profileDirs gets all the folders in the user's profiles folder regardless of valid or invalid config
func profileDirs(miniHome ...string) (dirs []string, err error) {
	miniPath := localpath.MiniPath()
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

// profileFilePath returns the Minikube profile config file
func profileFilePath(profile string, miniHome ...string) string {
	miniPath := localpath.MiniPath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}

	return filepath.Join(miniPath, "profiles", profile, "config.json")
}

// ProfileFolderPath returns path of profile folder
func ProfileFolderPath(profile string, miniHome ...string) string {
	miniPath := localpath.MiniPath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}
	return filepath.Join(miniPath, "profiles", profile)
}
