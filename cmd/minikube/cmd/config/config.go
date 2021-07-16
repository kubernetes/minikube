/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// Bootstrapper is the name for bootstrapper
const Bootstrapper = "bootstrapper"

type setFn func(string, string) error

// Setting represents a setting
type Setting struct {
	name          string
	description   string
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
		description:   "Used to specify the driver to run Kubernetes in. The list of available drivers depends on operating system.",
		set:           SetString,
		validDefaults: driver.SupportedDrivers,
		validations:   []setFn{IsValidDriver},
		callbacks:     []setFn{RequiresRestartMsg},
	},
	{
		name:        "vm-driver",
		description: "DEPRECATED, use driver instead.",
		set:         SetString,
		validations: []setFn{IsValidDriver},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "container-runtime",
		description: "The container runtime to be used (docker, cri-o, containerd). (default \"docker\")",
		set:         SetString,
		validations: []setFn{IsValidRuntime},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "feature-gates",
		description: "A set of key=value pairs that describe feature gates for alpha/experimental features.",
		set:         SetString,
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "v",
		description: "number for the log level verbosity",
		set:         SetInt,
		validations: []setFn{IsPositive},
	},
	{
		name:        "cpus",
		description: "Number of CPUs allocated to each minikube node",
		set:         SetInt,
		validations: []setFn{IsValidCPUs},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "disk-size",
		description: "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g). (default \"20000mb\")",
		set:         SetString,
		validations: []setFn{IsValidDiskSize},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "host-only-cidr",
		description: "The CIDR to be used for the minikube VM (virtualbox driver only) (default \"192.168.99.1/24\")",
		set:         SetString,
		validations: []setFn{IsValidCIDR},
	},
	{
		name:        "memory",
		description: "Amount of RAM to allocate to Kubernetes (format: <number>[<unit>], where unit = b, k, m or g).",
		set:         SetString,
		validations: []setFn{IsValidMemory},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "log_dir",
		description: "If non-empty, write log files in this directory",
		set:         SetString,
		validations: []setFn{IsValidPath},
	},
	{
		name:        "kubernetes-version",
		description: "The Kubernetes version that the minikube VM will use (ex: v1.2.3, 'stable' for v1.20.7, 'latest' for v1.22.0-alpha.2). Defaults to 'stable'.",
		set:         SetString,
	},
	{
		name:        "iso-url",
		description: fmt.Sprintf("Locations to fetch the minikube ISO from. (default %s)", download.DefaultISOURLs()),
		set:         SetString,
		validations: []setFn{IsValidURL, IsURLExists},
	},
	{
		name:        config.WantUpdateNotification,
		description: "If true, will notify on start if there's a newer stable version of minikube available.",
		set:         SetBool,
	},
	{
		name:        config.WantBetaUpdateNotification,
		description: "If true, will notify on start if there's a newer beta version of minikube available.",
		set:         SetBool,
	},
	{
		name:        config.ReminderWaitPeriodInHours,
		description: "The number of hours to wait before reminding there's a newer version of minikube again.",
		set:         SetInt,
	},
	{
		name:        config.WantNoneDriverWarning,
		description: "If false, will stop recommending to use Docker driver instead of none driver.",
		set:         SetBool,
	},
	{
		name:        config.ProfileName,
		description: "The name of the minikube cluster (profile) default \"minikube\"",
		set:         SetString,
	},
	{
		name:        Bootstrapper,
		description: "The name of the cluster bootstrapper that will set up the Kubernetes cluster. (default \"kubeadm\")",
		set:         SetString,
	},
	{
		name:        "insecure-registry",
		description: "Insecure image registries to pass to the container runtime engine (docker, containerd, cri-o). The default service CIDR range will automatically be added.",
		set:         SetString,
	},
	{
		name:        "hyperv-virtual-switch",
		description: "The hyperv virtual switch name. Defaults to first found. (hyperv driver only)",
		set:         SetString,
	},
	{
		name:        "disable-driver-mounts",
		description: "Disables the filesystem mounts provided by the hypervisors",
		set:         SetBool,
	},
	{
		name:   "cache",
		set:    SetConfigMap,
		setMap: SetMap,
	},
	{
		name:        config.EmbedCerts,
		description: "If true, will embed the certs in kubeconfig.",
		set:         SetBool,
	},
	{
		name:        "native-ssh",
		description: "Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'.",
		set:         SetBool,
	},
}

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config SUBCOMMAND [flags]",
	Short: "Modify persistent configuration values",
	Long: `config modifies minikube config files using subcommands like "minikube config set driver kvm2"
Configurable fields: ` + "\n\n" + configurableFields(),
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			klog.ErrorS(err, "help")
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
