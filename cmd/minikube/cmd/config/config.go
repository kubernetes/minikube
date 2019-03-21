/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// Bootstrapper is the name for bootstrapper
const Bootstrapper = "bootstrapper"

type setFn func(string, string) error

// Setting represents a setting
type Setting struct {
	name        string
	set         func(config.MinikubeConfig, string, string) error
	setMap      func(config.MinikubeConfig, string, map[string]interface{}) error
	validations []setFn
	callbacks   []setFn
}

// These are all the settings that are configurable
// and their validation and callback fn run on Set
var settings = []Setting{
	{
		name:        "vm-driver",
		set:         SetString,
		validations: []setFn{IsValidDriver},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:      "feature-gates",
		set:       SetString,
		callbacks: []setFn{RequiresRestartMsg},
	},
	{
		name:        "v",
		set:         SetInt,
		validations: []setFn{IsPositive},
	},
	{
		name:        "cpus",
		set:         SetInt,
		validations: []setFn{IsPositive},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "disk-size",
		set:         SetString,
		validations: []setFn{IsValidDiskSize},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "host-only-cidr",
		set:         SetString,
		validations: []setFn{IsValidCIDR},
	},
	{
		name:        "memory",
		set:         SetInt,
		validations: []setFn{IsPositive},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "log_dir",
		set:         SetString,
		validations: []setFn{IsValidPath},
	},
	{
		name: "kubernetes-version",
		set:  SetString,
	},
	{
		name:        "iso-url",
		set:         SetString,
		validations: []setFn{IsValidURL, IsURLExists},
	},
	{
		name: config.WantUpdateNotification,
		set:  SetBool,
	},
	{
		name: config.ReminderWaitPeriodInHours,
		set:  SetInt,
	},
	{
		name: config.WantReportError,
		set:  SetBool,
	},
	{
		name: config.WantReportErrorPrompt,
		set:  SetBool,
	},
	{
		name: config.WantKubectlDownloadMsg,
		set:  SetBool,
	},
	{
		name: config.WantNoneDriverWarning,
		set:  SetBool,
	},
	{
		name: config.MachineProfile,
		set:  SetString,
	},
	{
		name: Bootstrapper,
		set:  SetString, //TODO(r2d4): more validation here?
	},
	{
		name: config.ShowBootstrapperDeprecationNotification,
		set:  SetBool,
	},
	{
		name:        "dashboard",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "addon-manager",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "default-storageclass",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableStorageClasses},
	},
	{
		name:        "heapster",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "efk",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "ingress",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "registry",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "registry-creds",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "freshpod",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "default-storageclass",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableStorageClasses},
	},
	{
		name:        "storage-provisioner",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "storage-provisioner-gluster",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableStorageClasses},
	},
	{
		name:        "metrics-server",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "nvidia-driver-installer",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "nvidia-gpu-device-plugin",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:        "logviewer",
		set:         SetBool,
		validations: []setFn{IsValidAddon},
	},
	{
		name:        "gvisor",
		set:         SetBool,
		validations: []setFn{IsValidAddon, IsContainerdRuntime},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name: "hyperv-virtual-switch",
		set:  SetString,
	},
	{
		name: "disable-driver-mounts",
		set:  SetBool,
	},
	{
		name:   "cache",
		set:    SetConfigMap,
		setMap: SetMap,
	},
	{
		name: "embed-certs",
		set:  SetBool,
	},
}

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config SUBCOMMAND [flags]",
	Short: "Modify minikube config",
	Long: `config modifies minikube config files using subcommands like "minikube config set vm-driver kvm"
Configurable fields: ` + "\n\n" + configurableFields(),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func configurableFields() string {
	fields := []string{}
	for _, s := range settings {
		fields = append(fields, " * "+s.name)
	}
	return strings.Join(fields, "\n")
}

// ListConfigMap list entries from config file
func ListConfigMap(name string) ([]string, error) {
	configFile, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}
	var images []string
	if values, ok := configFile[name].(map[string]interface{}); ok {
		for key := range values {
			images = append(images, key)
		}
	}
	return images, nil
}

// AddToConfigMap adds entries to a map in the config file
func AddToConfigMap(name string, images []string) error {
	s, err := findSetting(name)
	if err != nil {
		return err
	}
	// Set the values
	configFile, err := config.ReadConfig()
	if err != nil {
		return err
	}
	newImages := make(map[string]interface{})
	for _, image := range images {
		newImages[image] = nil
	}
	if values, ok := configFile[name].(map[string]interface{}); ok {
		for key := range values {
			newImages[key] = nil
		}
	}
	if err = s.setMap(configFile, name, newImages); err != nil {
		return err
	}
	// Write the values
	return WriteConfig(configFile)
}

// DeleteFromConfigMap deletes entries from a map in the config file
func DeleteFromConfigMap(name string, images []string) error {
	s, err := findSetting(name)
	if err != nil {
		return err
	}
	// Set the values
	configFile, err := config.ReadConfig()
	if err != nil {
		return err
	}
	values, ok := configFile[name]
	if !ok {
		return nil
	}
	for _, image := range images {
		delete(values.(map[string]interface{}), image)
	}
	if err = s.setMap(configFile, name, values.(map[string]interface{})); err != nil {
		return err
	}
	// Write the values
	return WriteConfig(configFile)
}

// WriteConfig writes a minikube config to the JSON file
func WriteConfig(m config.MinikubeConfig) error {
	f, err := os.Create(constants.ConfigFile)
	if err != nil {
		return fmt.Errorf("Could not open file %s: %s", constants.ConfigFile, err)
	}
	defer f.Close()
	err = encode(f, m)
	if err != nil {
		return fmt.Errorf("Error encoding config %s: %s", constants.ConfigFile, err)
	}
	return nil
}

func encode(w io.Writer, m config.MinikubeConfig) error {
	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}
