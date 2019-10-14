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

package localpath

import (
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

const (
	// OverrideVar is the name of the minikube home directory variable.
	OverrideVar = "MINIKUBE_HOME"

	// Subdirectory to store data in.
	appDir = "minikube"
)

// CreateDirs creates directories, possibly migrating from an old disk layout
func CreateDirs() (map[string]string, error) {
	result := map[string]string{}
	for _, d := range []string{cacheDir(), configDir(), Addons(), FileSync()} {
		if _, err := os.Stat(d); err != nil {
			glog.Infof("mkdir: %s", d)
			if err := os.MkdirAll(d, 0700); err != nil {
				return result, errors.Wrap(err, "mkdir")
			}
			result[d] = ""
		}
	}

	oldHome := existingHome()
	if oldHome != "" {
		me, err := migrateLegacyPaths(oldHome)
		if err != nil {
			return result, errors.Wrap(err, "migrate")
		}
		for k, v := range me {
			result[k] = v
		}
	}
	return result, nil
}

// Machine return the location of a machine directory
func Machine(name string) string {
	return filepath.Join(Store(name), "machines", name)
}

// Store return the location of the libmachine store
func Store(name string) string {
	return filepath.Join(dataDir(), "stores", name)
}

// Profile return the location of a profile directory
func Profile(name string) string {
	return filepath.Join(Profiles(), name)
}

// Profile return the location of the profiles directory
func Profiles() string {
	return filepath.Join(configDir(), "profiles")
}

// ProfileConfig return the location of a profile configuration file
func ProfileConfig(name string) string {
	return filepath.Join(Profile(name), "config.json")
}

// MachineCerts returns the location of the machine certificates directory
func MachineCerts(name string) string {
	return filepath.Join(Store(name), "certs")
}

// KubernetesCerts returns the location of the Kubernetes certificates directory
func KubernetesCerts(name string) string {
	return filepath.Join(Profile(name), "kubernetes")
}

// SSHKey returns the location of the ssh key
func SSHKey(name string) string {
	return filepath.Join(Machine(name), "id_rsa")
}

// MountPid returns the path of the mount process pid
func MountPid(name string) string {
	return filepath.Join(Profile(name), "mount.pid")
}

// UpdateCheck returns the path of the update check file
func UpdateCheck() string {
	return filepath.Join(dataDir(), "last_update_check")
}

// ContainerImages returns the location of the container images directory
func ContainerImages() string {
	return filepath.Join(cacheDir(), "images")
}

// DiskImages returns the location of the disk images directory
func DiskImages() string {
	return filepath.Join(cacheDir(), "iso")
}

// Addons returns the location of the addons directory
func Addons() string {
	return filepath.Join(dataDir(), "addons")
}

// FileSync returns the location of the file sync directory
func FileSync() string {
	return filepath.Join(dataDir(), "files")
}

// Binaries returns the location of the binaries directory
func Binaries() string {
	return filepath.Join(cacheDir(), "binaries")
}

// Drivers returns the location of the drivers directory
func Drivers() string {
	return filepath.Join(dataDir(), "drivers")
}

// Logs returns the location of the logs directory
func Logs() string {
	return filepath.Join(dataDir(), "logs")
}

// TunnelRegistry returns the location of the tunnel registry
func TunnelRegistry() string {
	return filepath.Join(dataDir(), "tunnels.json")
}

// GlobalConfig returns the path to the global config file
func GlobalConfig() string {
	return filepath.Join(configDir(), "config.json")
}

func dataDir() string {
	if dir := overrideHome(); dir != "" {
		return filepath.Join(dir, "data")
	}
	// See https://github.com/golang/go/issues/29960
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, appDir)
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, appDir, "data")
	}
	return legacyHome()
}

func cacheDir() string {
	if dir := overrideHome(); dir != "" {
		return filepath.Join(dir, "cache")
	}

	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, appDir)
	}
	return filepath.Join(legacyHome(), "cache")
}

func configDir() string {
	if dir := overrideHome(); dir != "" {
		return filepath.Join(dir, "config")
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, appDir)
	}
	return legacyHome()
}

func overrideHome() string {
	val := os.Getenv(OverrideVar)
	if val == "" {
		return ""
	}
	if filepath.Base(val) == ".minikube" {
		return val
	}
	return filepath.Join(val, ".minikube")
}
