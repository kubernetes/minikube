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

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/util/lock"
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

// ProfileExists returns true if there is a profile config (regardless of being valid)
func ProfileExists(name string) bool {
	p := localpath.ProfileConfig(name)
	_, err := os.Stat(p)
	return err == nil
}

// CreateEmptyProfile creates an empty profile
func CreateEmptyProfile(name string) error {
	cfg := &Config{}
	return CreateProfile(name, cfg)
}

// CreateProfile creates an profile out of the cfg
func CreateProfile(name string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	path := localpath.ProfileConfig(name)
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

	if err = lock.WriteFile(tf.Name(), data, 0600); err != nil {
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
func DeleteProfile(profile string) error {
	return os.RemoveAll(localpath.Profile(profile))
}

// ListProfiles returns all valid and invalid (if any) minikube profiles
// invalidPs are the profiles that have a directory or config file but not usable
// invalidPs would be suggested to be deleted
func ListProfiles() (validPs []*Profile, inValidPs []*Profile, err error) {
	pDirs, err := profileDirs()
	if err != nil {
		return nil, nil, err
	}
	for _, n := range pDirs {
		p, err := loadProfile(n)
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
func loadProfile(name string) (*Profile, error) {
	cfg, err := DefaultLoader.LoadConfigFromFile(name)
	p := &Profile{
		Name:   name,
		Config: cfg,
	}
	return p, err
}

// profileDirs gets all the folders in the user's profiles folder regardless of valid or invalid config
func profileDirs() (dirs []string, err error) {
	items, err := ioutil.ReadDir(localpath.Profiles())
	for _, f := range items {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}
	return dirs, err
}
