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

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/registry"
)

const (
	Mock         = "mock"
	None         = "none"
	KVM2         = "kvm2"
	VirtualBox   = "virtualbox"
	HyperKit     = "hyperkit"
	VMware       = "vmware"
	VMwareFusion = "vmwarefusion"
	HyperV       = "hyperv"
	Parallels    = "parallels"
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
		return FlagHints{}
	}

	// for more info see: https://github.com/kubernetes/minikube/issues/3511
	f := "/run/systemd/resolve/resolv.conf"
	extraOpts := ""
	if _, err := os.Stat(f); err == nil {
		extraOpts = fmt.Sprintf("kubelet.resolv-conf=%s", f)
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
			glog.Warningf("%s is installed, but unhealthy: %v", ds.Name, ds.State.Error)
			continue
		}
		options = append(options, ds)
		glog.Infof("%q driver appears healthy, priority %d", ds.Name, ds.Priority)

	}
	return options
}

// Choose returns a suggested driver from a set of options
func Choose(options []registry.DriverState) (registry.DriverState, []registry.DriverState) {
	pick := registry.DriverState{}
	for _, ds := range options {
		if ds.Priority > pick.Priority {
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
