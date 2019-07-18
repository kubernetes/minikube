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
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/constants"
)

// AllProfiles returns all minikube profiles
func AllProfiles() (profiles []*Profile) {

	pDirs, err := allProfileDirs()
	if err != nil {
		glog.Infof("Error getting profile directories %v", err)
	}
	for _, n := range pDirs {
		p, err := loadProfile(n)
		if err != nil {
			// TODO warn user to delete this folder or maybe do it for them
			glog.Errorf("Invalid profile config in profiles folder (%s):\n error: %v", n, err)
			continue
		}
		profiles = append(profiles, p)
	}
	return profiles
}

func loadProfile(n string) (*Profile, error) {

	cfg, err := DefaultLoader.LoadConfigFromFile(n)
	profile := &Profile{
		Name:   n,
		Config: cfg,
	}
	return profile, err
}

// allProfileDirs gets all the folders in the user's profiles folder
func allProfileDirs() (dirs []string, err error) {
	root := filepath.Join(constants.GetMinipath(), "profiles")
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if path != root {
				dirs = append(dirs, filepath.Base(path))
			}
		}
		return nil
	})
	return dirs, err

}
