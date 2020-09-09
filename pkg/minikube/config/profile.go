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
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/util/lock"
)

var keywords = []string{"start", "stop", "status", "delete", "config", "open", "profile", "addons", "cache", "logs"}

// IsValid checks if the profile has the essential info needed for a profile
func (p *Profile) IsValid() bool {
	if p.Config == nil {
		return false
	}
	if p.Config == nil {
		return false
	}
	if p.Config.Driver == "" {
		return false
	}
	for _, n := range p.Config.Nodes {
		if n.KubernetesVersion == "" {
			return false
		}
	}
	return true
}

// PrimaryControlPlane gets the node specific config for the first created control plane
func PrimaryControlPlane(cc *ClusterConfig, miniHome ...string) (Node, error) {
	miniPath := localpath.MiniPath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}

	for _, n := range cc.Nodes {
		if n.ControlPlane {
			return n, nil
		}
	}

	// This config is probably from 1.6 or earlier, let's convert it.
	cp := Node{
		Name:              cc.KubernetesConfig.NodeName,
		IP:                cc.KubernetesConfig.NodeIP,
		Port:              cc.KubernetesConfig.NodePort,
		KubernetesVersion: cc.KubernetesConfig.KubernetesVersion,
		ControlPlane:      true,
		Worker:            true,
	}

	cc.Nodes = []Node{cp}

	// Remove old style attribute to avoid confusion
	cc.KubernetesConfig.NodeName = ""
	cc.KubernetesConfig.NodeIP = ""

	err := SaveProfile(viper.GetString(ProfileName), cc, miniPath)
	if err != nil {
		return Node{}, err
	}

	return cp, nil
}

// ProfileNameValid checks if the profile name is container name and DNS hostname/label friendly.
func ProfileNameValid(name string) bool {
	// RestrictedNamePattern describes the characters allowed to represent a profile's name
	const RestrictedNamePattern = `(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])`

	var validName = regexp.MustCompile(`^` + RestrictedNamePattern + `$`)

	return validName.MatchString(name)
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

// CreateEmptyProfile creates an empty profile and stores in $MINIKUBE_HOME/profiles/<profilename>/config.json
func CreateEmptyProfile(name string, miniHome ...string) error {
	cfg := &ClusterConfig{}
	return SaveProfile(name, cfg, miniHome...)
}

// SaveNode saves a node to a cluster
func SaveNode(cfg *ClusterConfig, node *Node) error {
	update := false
	for i, n := range cfg.Nodes {
		if n.Name == node.Name {
			cfg.Nodes[i] = *node
			update = true
			break
		}
	}

	if !update {
		cfg.Nodes = append(cfg.Nodes, *node)
	}

	return SaveProfile(viper.GetString(ProfileName), cfg)
}

// SaveProfile creates an profile out of the cfg and stores in $MINIKUBE_HOME/profiles/<profilename>/config.json
func SaveProfile(name string, cfg *ClusterConfig, miniHome ...string) error {
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

	// try to get profiles list based on left over evidences such as directory
	pDirs, err := profileDirs(miniHome...)
	if err != nil {
		return nil, nil, err
	}
	// try to get profiles list based on all contrainers created by docker driver
	cs, err := oci.ListOwnedContainers(oci.Docker)
	if err == nil {
		pDirs = append(pDirs, cs...)
	}
	pDirs = removeDupes(pDirs)
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

// removeDupes removes duplicates
func removeDupes(profiles []string) []string {
	// Use map to record duplicates as we find them.
	seen := map[string]bool{}
	result := []string{}

	for n := range profiles {
		if seen[profiles[n]] {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			seen[profiles[n]] = true
			// Append to result slice.
			result = append(result, profiles[n])
		}
	}
	// Return the new slice.
	return result
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

// profileFilePath returns path of profile config file
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
