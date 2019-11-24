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
	"io/ioutil"
	"os"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/localpath"
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

var (
	// ErrKeyNotFound is the error returned when a key doesn't exist in the config file
	ErrKeyNotFound = errors.New("specified key could not be found in config")
)

// MinikubeConfig represents minikube config
type MinikubeConfig map[string]interface{}

// Get gets a named value from config
func Get(name string) (string, error) {
	m, err := ReadConfig(localpath.ConfigFile)
	if err != nil {
		return "", err
	}
	return get(name, m)
}

func get(name string, config MinikubeConfig) (string, error) {
	if val, ok := config[name]; ok {
		return fmt.Sprintf("%v", val), nil
	}
	return "", ErrKeyNotFound
}

// WriteConfig writes a minikube config to the JSON file
func WriteConfig(configFile string, m MinikubeConfig) error {
	f, err := os.Create(configFile)
	if err != nil {
		return fmt.Errorf("create %s: %s", configFile, err)
	}
	defer f.Close()
	err = encode(f, m)
	if err != nil {
		return fmt.Errorf("encode %s: %s", configFile, err)
	}
	return nil
}

// ReadConfig reads in the JSON minikube config
func ReadConfig(configFile string) (MinikubeConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("open %s: %v", localpath.ConfigFile, err)
	}
	defer f.Close()

	m, err := decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %v", localpath.ConfigFile, err)
	}

	return m, nil
}

func decode(r io.Reader) (MinikubeConfig, error) {
	var data MinikubeConfig
	err := json.NewDecoder(r).Decode(&data)
	return data, err
}

func encode(w io.Writer, m MinikubeConfig) error {
	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}

// Load loads the kubernetes and machine config for the current machine
func Load() (*MachineConfig, error) {
	machine := viper.GetString(MachineProfile)
	return DefaultLoader.LoadConfigFromFile(machine)
}

// Loader loads the kubernetes and machine config based on the machine profile name
type Loader interface {
	LoadConfigFromFile(profile string, miniHome ...string) (*MachineConfig, error)
}

type simpleConfigLoader struct{}

// DefaultLoader is the default config loader
var DefaultLoader Loader = &simpleConfigLoader{}

func (c *simpleConfigLoader) LoadConfigFromFile(profileName string, miniHome ...string) (*MachineConfig, error) {
	var cc MachineConfig
	// Move to profile package
	path := profileFilePath(profileName, miniHome...)

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
