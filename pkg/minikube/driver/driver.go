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

package driver

import (
	"fmt"
	"os"
	"sort"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/registry"
)

const (
	// Mock driver
	Mock = "mock"
	// None driver
	None = "none"
	// KVM2 driver
	KVM2 = "kvm2"
	// VirtualBox driver
	VirtualBox = "virtualbox"
	// HyperKit driver
	HyperKit = "hyperkit"
	// VMware driver
	VMware = "vmware"
	// VMwareFusion driver
	VMwareFusion = "vmwarefusion"
	// HyperV driver
	HyperV = "hyperv"
	// Parallels driver
	Parallels = "parallels"
)

var (
	// systemdResolvConf is path to systemd's DNS configuration. https://github.com/kubernetes/minikube/issues/3511
	systemdResolvConf = "/run/systemd/resolve/resolv.conf"
)

// SupportedDrivers returns a list of supported drivers
func SupportedDrivers() []string {
	return supportedDrivers
}

// Supported returns if the driver is supported on this host.
func Supported(name string) bool {
	for _, d := range supportedDrivers {
		if name == d {
			return true
		}
	}
	return false
}

// BareMetal returns if this driver is unisolated
func BareMetal(name string) bool {
	return name == None || name == Mock
}

// FlagHints are hints for what default options should be used for this driver
type FlagHints struct {
	ExtraOptions string
	CacheImages  bool
}

// FlagDefaults returns suggested defaults based on a driver
func FlagDefaults(name string) FlagHints {
	if name != None {
		return FlagHints{CacheImages: true}
	}

	extraOpts := ""
	if _, err := os.Stat(systemdResolvConf); err == nil {
		extraOpts = fmt.Sprintf("kubelet.resolv-conf=%s", systemdResolvConf)
	}
	return FlagHints{
		ExtraOptions: extraOpts,
		CacheImages:  false,
	}
}

// Choices returns a list of drivers which are possible on this system
func Choices() []registry.DriverState {
	options := []registry.DriverState{}
	for _, ds := range registry.Installed() {
		if !ds.State.Healthy {
			glog.Warningf("%q is installed, but unhealthy: %v", ds.Name, ds.State.Error)
			continue
		}
		options = append(options, ds)
	}

	// Descending priority for predictability and appearance
	sort.Slice(options, func(i, j int) bool {
		return options[i].Priority > options[j].Priority
	})
	return options
}

// Choose returns a suggested driver from a set of options
func Choose(options []registry.DriverState) (registry.DriverState, []registry.DriverState) {
	pick := registry.DriverState{}
	for _, ds := range options {
		if ds.Priority <= registry.Discouraged {
			glog.Infof("not recommending %q due to priority: %d", ds.Name, ds.Priority)
			continue
		}
		if ds.Priority > pick.Priority {
			glog.V(1).Infof("%q has a higher priority (%d) than %q (%d)", ds.Name, ds.Priority, pick.Name, pick.Priority)
			pick = ds
		}
	}

	alternates := []registry.DriverState{}
	for _, ds := range options {
		if ds != pick {
			alternates = append(alternates, ds)
		}
	}
	glog.Infof("Picked: %+v", pick)
	glog.Infof("Alternatives: %+v", alternates)
	return pick, alternates
}

// Status returns the status of a driver
func Status(name string) registry.State {
	return registry.Status(name)
}
