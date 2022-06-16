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
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
)

// IsValidDriver checks if a driver is supported
func IsValidDriver(string, name string) error {
	if driver.Supported(name) {
		return nil
	}
	return fmt.Errorf("driver %q is not supported", name)
}

// RequiresRestartMsg returns the "requires restart" message
func RequiresRestartMsg(string, string) error {
	out.WarningT("These changes will take effect upon a minikube delete and then a minikube start")
	return nil
}

// IsValidDiskSize checks if a string is a valid disk size
func IsValidDiskSize(name string, disksize string) error {
	_, err := units.FromHumanSize(disksize)
	if err != nil {
		return fmt.Errorf("invalid disk size: %v", err)
	}
	return nil
}

// IsValidCPUs checks if a string is a valid number of CPUs
func IsValidCPUs(name string, cpus string) error {
	if cpus == constants.MaxResources {
		return nil
	}
	return IsPositive(name, cpus)
}

// IsValidMemory checks if a string is a valid memory size
func IsValidMemory(name string, memsize string) error {
	if memsize == constants.MaxResources {
		return nil
	}
	_, err := units.FromHumanSize(memsize)
	if err != nil {
		return fmt.Errorf("invalid memory size: %v", err)
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
		return fmt.Errorf("invalid CIDR: %v", err)
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

// IsValidRuntime checks if a string is a valid runtime
func IsValidRuntime(name string, runtime string) error {
	_, err := cruntime.New(cruntime.Config{Type: runtime})
	if err != nil {
		return fmt.Errorf("invalid runtime: %v", err)
	}
	return nil
}
