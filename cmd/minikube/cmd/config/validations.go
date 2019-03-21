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
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	units "github.com/docker/go-units"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

// IsValidDriver checks if a driver is supported
func IsValidDriver(string, driver string) error {
	for _, d := range constants.SupportedVMDrivers {
		if driver == d {
			return nil
		}
	}
	return fmt.Errorf("Driver %s is not supported", driver)
}

// RequiresRestartMsg returns the "requires restart" message
func RequiresRestartMsg(string, string) error {
	console.OutStyle("warning", "These changes will take effect upon a minikube delete and then a minikube start")
	return nil
}

// IsValidDiskSize checks if a string is a valid disk size
func IsValidDiskSize(name string, disksize string) error {
	_, err := units.FromHumanSize(disksize)
	if err != nil {
		return fmt.Errorf("Not valid disk size: %v", err)
	}
	return nil
}

// IsValidURL checks if a location is a valid URL
func IsValidURL(name string, location string) error {
	_, err := url.Parse(location)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL", location)
	}
	return nil
}

// IsURLExists checks if a location actually exists
func IsURLExists(name string, location string) error {
	parsed, err := url.Parse(location)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL", location)
	}

	// we can only validate if local files exist, not other urls
	if parsed.Scheme != "file" {
		return nil
	}

	// chop off "file://" from the location, giving us the real system path
	sysPath := strings.TrimPrefix(location, "file://")
	stat, err := os.Stat(sysPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", location)
		}

		if os.IsPermission(err) {
			return fmt.Errorf("%s could not be opened (permission error: %s)", location, err.Error())
		}

		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is a directory", location)
	}

	return nil
}

// IsPositive checks if an integer is positive
func IsPositive(name string, val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("%s:%v", name, err)
	}
	if i <= 0 {
		return fmt.Errorf("%s must be > 0", name)
	}
	return nil
}

// IsValidCIDR checks if a string parses as a CIDR
func IsValidCIDR(name string, cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("Error parsing CIDR: %v", err)
	}
	return nil
}

// IsValidPath checks if a string is a valid path
func IsValidPath(name string, path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s path is not valid: %v", name, err)
	}
	return nil
}

// IsValidAddon checks if a string is a valid addon
func IsValidAddon(name string, val string) error {
	if _, ok := assets.Addons[name]; ok {
		return nil
	}
	return errors.Errorf("Cannot enable/disable invalid addon %s", name)
}

// IsContainerdRuntime is a validator which returns an error if the current runtime is not containerd
func IsContainerdRuntime(_, _ string) error {
	config, err := config.Load()
	if err != nil {
		return fmt.Errorf("config.Load: %v", err)
	}
	r, err := cruntime.New(cruntime.Config{Type: config.KubernetesConfig.ContainerRuntime})
	if err != nil {
		return err
	}
	_, ok := r.(*cruntime.Containerd)
	if !ok {
		return fmt.Errorf(`This addon can only be enabled with the containerd runtime backend.

To enable this backend, please first stop minikube with:

minikube stop

and then start minikube again with the following flags:

minikube start --container-runtime=containerd --docker-opt containerd=/var/run/containerd/containerd.sock`)
	}

	return nil
}
