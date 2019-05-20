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
	"errors"
	"fmt"
	"io"
	"os"

	"io/ioutil"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/constants"
)

const (
	// WantUpdateNotification is the key for WantUpdateNotification
	WantUpdateNotification = "WantUpdateNotification"
	// ReminderWaitPeriodInHours is the key for WantUpdateNotification
	ReminderWaitPeriodInHours = "ReminderWaitPeriodInHours"
	// WantReportError is the key for WantReportError
	WantReportError = "WantReportError"
	// WantReportErrorPrompt is the key for WantReportErrorPrompt
	WantReportErrorPrompt = "WantReportErrorPrompt"
	// WantKubectlDownloadMsg is the key for WantKubectlDownloadMsg
	WantKubectlDownloadMsg = "WantKubectlDownloadMsg"
	// WantNoneDriverWarning is the key for WantNoneDriverWarning
	WantNoneDriverWarning = "WantNoneDriverWarning"
	// MachineProfile is the key for MachineProfile
	MachineProfile = "profile"
	// ShowDriverDeprecationNotification is the key for ShowDriverDeprecationNotification
	ShowDriverDeprecationNotification = "ShowDriverDeprecationNotification"
	// ShowBootstrapperDeprecationNotification is the key for ShowBootstrapperDeprecationNotification
	ShowBootstrapperDeprecationNotification = "ShowBootstrapperDeprecationNotification"
)

// MinikubeConfig represents minikube config
type MinikubeConfig map[string]interface{}

// Get gets a named value from config
func Get(name string) (string, error) {
	m, err := ReadConfig()
	if err != nil {
		return "", err
	}
	return get(name, m)
}

func get(name string, config MinikubeConfig) (string, error) {
	if val, ok := config[name]; ok {
		return fmt.Sprintf("%v", val), nil
	}
	return "", errors.New("specified key could not be found in config")
}

// ReadConfig reads in the JSON minikube config
func ReadConfig() (MinikubeConfig, error) {
	f, err := os.Open(constants.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("open %s: %v", constants.ConfigFile, err)
	}
	defer f.Close()

	m, err := decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %v", constants.ConfigFile, err)
	}

	return m, nil
}

func decode(r io.Reader) (MinikubeConfig, error) {
	var data MinikubeConfig
	err := json.NewDecoder(r).Decode(&data)
	return data, err
}

// GetMachineName gets the machine name for the VM
func GetMachineName() string {
	if viper.GetString(MachineProfile) == "" {
		return constants.DefaultMachineName
	}
	return viper.GetString(MachineProfile)
}

// Load loads the kubernetes and machine config for the current machine
func Load() (*Config, error) {
	return DefaultLoader.LoadConfigFromFile(GetMachineName())
}

// Loader loads the kubernetes and machine config based on the machine profile name
type Loader interface {
	LoadConfigFromFile(profile string) (*Config, error)
}

type simpleConfigLoader struct{}

// DefaultLoader is the default config loader
var DefaultLoader Loader = &simpleConfigLoader{}

func (c *simpleConfigLoader) LoadConfigFromFile(profile string) (*Config, error) {
	var cc Config

	path := constants.GetProfileFile(profile)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, err
	}
	return &cc, nil
}
