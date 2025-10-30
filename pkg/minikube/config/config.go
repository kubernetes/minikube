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

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/localpath"
)

const (
	// WantUpdateNotification is the key for WantUpdateNotification
	WantUpdateNotification = "WantUpdateNotification"
	// WantBetaUpdateNotification is the key for WantBetaUpdateNotification
	WantBetaUpdateNotification = "WantBetaUpdateNotification"
	// ReminderWaitPeriodInHours is the key for ReminderWaitPeriodInHours
	ReminderWaitPeriodInHours = "ReminderWaitPeriodInHours"
	// WantNoneDriverWarning is the key for WantNoneDriverWarning
	WantNoneDriverWarning = "WantNoneDriverWarning"
	// WantVirtualBoxDriverWarning is the key for WantVirtualBoxDriverWarning
	WantVirtualBoxDriverWarning = "WantVirtualBoxDriverWarning"
	// ProfileName represents the key for the global profile parameter
	ProfileName = "profile"
	// UserFlag is the key for the global user flag (ex. --user=user1)
	UserFlag = "user"
	// SkipAuditFlag is the key for skipping command from aduit
	SkipAuditFlag = "skip-audit"
	// Rootless is the key for the global rootless parameter (boolean)
	Rootless = "rootless"
	// AddonImages stores custom addon images config
	AddonImages = "addon-images"
	// AddonRegistries stores custom addon images config
	AddonRegistries = "addon-registries"
	// AddonListFlag represents the key for addons parameter
	AddonListFlag = "addons"
	// EmbedCerts represents the config for embedding certificates in kubeconfig
	EmbedCerts = "EmbedCerts"
	// MaxAuditEntries is the maximum number of audit entries to retain
	MaxAuditEntries = "MaxAuditEntries"
	// DefaultDnsDomain is the key for the default dns domain name when creating
	// a cluster
	DefaultDNSDomain = "default-dns-domain"
)

var (
	// ErrKeyNotFound is the error returned when a key doesn't exist in the config file
	ErrKeyNotFound = errors.New("specified key could not be found in config")
	// DockerEnv contains the environment variables
	DockerEnv []string
	// DockerOpt contains the option parameters
	DockerOpt []string
	// ExtraOptions contains extra options (if any)
	ExtraOptions ExtraOptionSlice
)

// ErrNotExist is the error returned when a config does not exist
type ErrNotExist struct {
	s string
}

func (e *ErrNotExist) Error() string {
	return e.s
}

// IsNotExist returns whether the error means a nonexistent configuration
func IsNotExist(err error) bool {
	if _, ok := err.(*ErrNotExist); ok {
		return true
	}
	return false
}

// ErrPermissionDenied is the error returned when the config cannot be read
// due to insufficient permissions
type ErrPermissionDenied struct {
	s string
}

func (e *ErrPermissionDenied) Error() string {
	return e.s
}

// IsPermissionDenied returns whether the error is a ErrPermissionDenied instance
func IsPermissionDenied(err error) bool {
	if _, ok := err.(*ErrPermissionDenied); ok {
		return true
	}
	return false
}

// MinikubeConfig represents minikube config
type MinikubeConfig map[string]interface{}

// Get gets a named value from config
func Get(name string) (string, error) {
	m, err := ReadConfig(localpath.ConfigFile())
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
		return nil, fmt.Errorf("open %s: %v", localpath.ConfigFile(), err)
	}
	defer f.Close()

	m, err := decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %v", localpath.ConfigFile(), err)
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

// Load loads the Kubernetes and machine config for the current machine
func Load(profile string, miniHome ...string) (*ClusterConfig, error) {
	cc, err := DefaultLoader.LoadConfigFromFile(profile, miniHome...)
	if err == nil {
		klog.Infof("Loaded profile config \"%s\": Driver=%s, ContainerRuntime=%s, KubernetesVersion=%s",
			profile,
			cc.Driver,
			cc.KubernetesConfig.ContainerRuntime,
			cc.KubernetesConfig.KubernetesVersion)
	}
	return cc, err
}

// Write writes the Kubernetes and machine config for the current machine
func Write(profile string, cc *ClusterConfig) error {
	return DefaultLoader.WriteConfigToFile(profile, cc)
}

// Loader loads the Kubernetes and machine config based on the machine profile name
type Loader interface {
	LoadConfigFromFile(profile string, miniHome ...string) (*ClusterConfig, error)
	WriteConfigToFile(profileName string, cc *ClusterConfig, miniHome ...string) error
}

type simpleConfigLoader struct{}

// DefaultLoader is the default config loader
var DefaultLoader Loader = &simpleConfigLoader{}

func (c *simpleConfigLoader) LoadConfigFromFile(profileName string, miniHome ...string) (*ClusterConfig, error) {
	var cc ClusterConfig
	// Move to profile package
	path := profileFilePath(profileName, miniHome...)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, &ErrNotExist{fmt.Sprintf("cluster %q does not exist", profileName)}
		}
		return nil, errors.Wrap(err, "stat")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsPermission(err) {
			return nil, &ErrPermissionDenied{err.Error()}
		}
		return nil, errors.Wrap(err, "read")
	}

	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return &cc, nil
}

func (c *simpleConfigLoader) WriteConfigToFile(profileName string, cc *ClusterConfig, miniHome ...string) error {
	path := profileFilePath(profileName, miniHome...)
	contents, err := json.MarshalIndent(cc, "", "	")
	if err != nil {
		return err
	}
	return os.WriteFile(path, contents, 0644)
}

// MultiNode returns true if the cluster has multiple nodes or if the request is asking for multinode
func MultiNode(cc ClusterConfig) bool {
	if len(cc.Nodes) > 1 {
		return true
	}
	return viper.GetInt("nodes") > 1
}

// IsHA returns true if ha (multi-control plane) cluster is requested.
func IsHA(cc ClusterConfig) bool {
	if len(ControlPlanes(cc)) > 1 {
		return true
	}
	return viper.GetBool("ha")
}
