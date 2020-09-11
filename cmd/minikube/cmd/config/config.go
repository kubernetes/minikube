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
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// Bootstrapper is the name for bootstrapper
const Bootstrapper = "bootstrapper"

type setFn func(string, string) error

// Setting represents a setting
type Setting struct {
	name          string
	set           func(config.MinikubeConfig, string, string) error
	setMap        func(config.MinikubeConfig, string, map[string]interface{}) error
	validDefaults func() []string
	validations   []setFn
	callbacks     []setFn
}

// These are all the settings that are configurable
// and their validation and callback fn run on Set
var settings = []Setting{
	{
		name:          "driver",
		set:           SetString,
		validDefaults: driver.SupportedDrivers,
		validations:   []setFn{IsValidDriver},
		callbacks:     []setFn{RequiresRestartMsg},
	},
	{
		name:        "vm-driver",
		set:         SetString,
		validations: []setFn{IsValidDriver},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "container-runtime",
		set:         SetString,
		validations: []setFn{IsValidRuntime},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:      "feature-gates",
		set:       SetString,
		callbacks: []setFn{RequiresRestartMsg},
	},
	{
		name:      "logging-format",
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
		name: config.ProfileName,
		set:  SetString,
	},
	{
		name: Bootstrapper,
		set:  SetString, //TODO(r2d4): more validation here?
	},
	{
		name: config.ShowDriverDeprecationNotification,
		set:  SetBool,
	},
	{
		name: config.ShowBootstrapperDeprecationNotification,
		set:  SetBool,
	},
	{
		name: "insecure-registry",
		set:  SetString,
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
	{
		name: "native-ssh",
		set:  SetBool,
	},
}

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config SUBCOMMAND [flags]",
	Short: "Modify persistent configuration values",
	Long: `config modifies minikube config files using subcommands like "minikube config set driver kvm"
Configurable fields: ` + "\n\n" + configurableFields(),
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			glog.Errorf("help: %v", err)
		}
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
	configFile, err := config.ReadConfig(localpath.ConfigFile())
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
	cfg, err := config.ReadConfig(localpath.ConfigFile())
	if err != nil {
		return err
	}
	newImages := make(map[string]interface{})
	for _, image := range images {
		newImages[image] = nil
	}
	if values, ok := cfg[name].(map[string]interface{}); ok {
		for key := range values {
			newImages[key] = nil
		}
	}
	if err = s.setMap(cfg, name, newImages); err != nil {
		return err
	}
	// Write the values
	return config.WriteConfig(localpath.ConfigFile(), cfg)
}

// DeleteFromConfigMap deletes entries from a map in the config file
func DeleteFromConfigMap(name string, images []string) error {
	s, err := findSetting(name)
	if err != nil {
		return err
	}
	// Set the values
	cfg, err := config.ReadConfig(localpath.ConfigFile())
	if err != nil {
		return err
	}
	values, ok := cfg[name]
	if !ok {
		return nil
	}
	for _, image := range images {
		delete(values.(map[string]interface{}), image)
	}
	if err = s.setMap(cfg, name, values.(map[string]interface{})); err != nil {
		return err
	}
	// Write the values
	return config.WriteConfig(localpath.ConfigFile(), cfg)
}
