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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/util/lock"
)

var keywords = []string{"start", "stop", "status", "delete", "config", "open", "profile", "addons", "cache", "logs"}

// IsValid checks if the profile has the essential info needed for a profile
func (p *Profile) IsValid() bool {
	if p.Config == nil {
		return false
	}
	if len(p.Config) == 0 {
		return false
	}
	// This will become a loop for multinode
	if p.Config[0] == nil {
		return false
	}
	if p.Config[0].VMDriver == "" {
		return false
	}
	if p.Config[0].KubernetesConfig.KubernetesVersion == "" {
		return false
	}
	return true
}

// ProfileNameInReservedKeywords checks if the profile is an internal keywords
func ProfileNameInReservedKeywords(name string) bool {
	for _, v := range keywords {
		if strings.EqualFold(v, name) {
			return true
		}
	}
	return false
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

// CreateEmptyProfile creates an empty profile stores in $MINIKUBE_HOME/profiles/<profilename>/config.json
func CreateEmptyProfile(name string, miniHome ...string) error {
	cfg := &MachineConfig{}
	return CreateProfile(name, cfg, miniHome...)
}

// CreateProfile creates an profile out of the cfg and stores in $MINIKUBE_HOME/profiles/<profilename>/config.json
func CreateProfile(name string, cfg *MachineConfig, miniHome ...string) error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	path := profileFilePath(name, miniHome...)
	glog.Infof("Saving config to %s ...", path)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	// If no config file exists, don't worry about swapping paths
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := lock.WriteFile(path, data, 0600); err != nil {
			return err
		}
		return nil
	}

	tf, err := ioutil.TempFile(filepath.Dir(path), "config.json.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tf.Name())

	if err = ioutil.WriteFile(tf.Name(), data, 0600); err != nil {
		return err
	}

	if err = tf.Close(); err != nil {
		return err
	}

	if err = os.Remove(path); err != nil {
		return err
	}

	if err = os.Rename(tf.Name(), path); err != nil {
		return err
	}
	return nil
}

// DeleteProfile deletes a profile and removes the profile dir
func DeleteProfile(profile string, miniHome ...string) error {
	miniPath := localpath.MiniPath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}
	return os.RemoveAll(ProfileFolderPath(profile, miniPath))
}

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
		Config: []*MachineConfig{cfg},
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
