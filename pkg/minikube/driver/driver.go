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
	"runtime"
	"sort"
	"strings"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/registry"
)

const (
	// Podman is Kubernetes in container using podman driver
	Podman = "podman"
	// Docker is Kubernetes in container using docker driver
	Docker = "docker"
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

// DisplaySupportedDrivers returns a string with a list of supported drivers
func DisplaySupportedDrivers() string {
	var sd []string
	for _, d := range supportedDrivers {
		if registry.Driver(d).Priority == registry.Experimental {
			sd = append(sd, d+" (experimental)")
			continue
		}
		sd = append(sd, d)
	}
	return strings.Join(sd, ", ")
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

// MachineType returns appropriate machine name for the driver
func MachineType(name string) string {
	if IsKIC(name) {
		return "container"
	}

	if IsVM(name) {
		return "VM"
	}

	// none or mock
	return "bare metal machine"
}

// IsKIC checks if the driver is a Kubernetes in container
func IsKIC(name string) bool {
	return name == Docker || name == Podman
}

// IsDocker checks if the driver docker
func IsDocker(name string) bool {
	return name == Docker
}

// IsKIC checks if the driver is a Docker for Desktop (Docker on windows or MacOs)
// for linux and exotic archs, this will be false
func IsDockerDesktop(name string) bool {
	if IsDocker(name) {
		if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
			return true
		}
	}
	return false
}

// IsMock checks if the driver is a mock
func IsMock(name string) bool {
	return name == Mock
}

// IsVM checks if the driver is a VM
func IsVM(name string) bool {
	if IsKIC(name) || BareMetal(name) {
		return false
	}
	return true
}

// BareMetal returns if this driver is unisolated
func BareMetal(name string) bool {
	return name == None || name == Mock
}

// NeedsRoot returns true if driver needs to run with root privileges
func NeedsRoot(name string) bool {
	return name == None
}

// NeedsPortForward returns true if driver is unable provide direct IP connectivity
func NeedsPortForward(name string) bool {
	// Docker for Desktop
	return IsKIC(name) && (runtime.GOOS == "darwin" || runtime.GOOS == "windows" || IsMicrosoftWSL())
}

// IsMicrosoftWSL will return true if process is running in WSL in windows
// checking for WSL env var based on this https://github.com/microsoft/WSL/issues/423#issuecomment-608237689
// also based on https://github.com/microsoft/vscode/blob/90a39ba0d49d75e9a4d7e62a6121ad946ecebc58/resources/win32/bin/code.sh#L24
func IsMicrosoftWSL() bool {
	return os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSLPATH") != "" || os.Getenv("WSLENV") != ""
}

// HasResourceLimits returns true if driver can set resource limits such as memory size or CPU count.
func HasResourceLimits(name string) bool {
	return name != None
}

// NeedsShutdown returns true if driver needs manual shutdown command before stopping.
// Hyper-V requires special care to avoid ACPI and file locking issues
// KIC also needs shutdown to avoid container getting stuck, https://github.com/kubernetes/minikube/issues/7657
func NeedsShutdown(name string) bool {
	if name == HyperV || IsKIC(name) {
		return true
	}
	return false
}

// FullName will return the human-known and title formatted name for the driver based on platform
func FullName(name string) string {
	switch name {
	case oci.Docker:
		if IsDockerDesktop(name) {
			return "Docker for Desktop"
		}
		return "Docker Service"
	default:
		return strings.Title(name)
	}
}

// FlagHints are hints for what default options should be used for this driver
type FlagHints struct {
	ExtraOptions     []string
	CacheImages      bool
	ContainerRuntime string
	Bootstrapper     string
}

// FlagDefaults returns suggested defaults based on a driver
func FlagDefaults(name string) FlagHints {
	fh := FlagHints{}
	if name != None {
		fh.CacheImages = true
		return fh
	}

	fh.CacheImages = false
	// if specifc linux add this option for systemd work on none driver
	if _, err := os.Stat(systemdResolvConf); err == nil {
		noneEO := fmt.Sprintf("kubelet.resolv-conf=%s", systemdResolvConf)
		fh.ExtraOptions = append(fh.ExtraOptions, noneEO)
		return fh
	}
	return fh
}

// Choices returns a list of drivers which are possible on this system
func Choices(vm bool) []registry.DriverState {
	options := registry.Available(vm)

	// Descending priority for predictability and appearance
	sort.Slice(options, func(i, j int) bool {
		return options[i].Priority > options[j].Priority
	})
	return options
}

// Suggest returns a suggested driver, alternate drivers, and rejected drivers
func Suggest(options []registry.DriverState) (registry.DriverState, []registry.DriverState, []registry.DriverState) {
	pick := registry.DriverState{}
	for _, ds := range options {
		if !ds.State.Installed {
			continue
		}

		if !ds.State.Healthy {
			glog.Infof("not recommending %q due to health: %v", ds.Name, ds.State.Error)
			continue
		}

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
	rejects := []registry.DriverState{}
	for _, ds := range options {
		if ds != pick {
			if !ds.State.Installed {
				ds.Rejection = fmt.Sprintf("Not installed: %v", ds.State.Error)
				rejects = append(rejects, ds)
				continue
			}

			if !ds.State.Healthy {
				ds.Rejection = fmt.Sprintf("Not healthy: %v", ds.State.Error)
				rejects = append(rejects, ds)
				continue
			}

			ds.Rejection = fmt.Sprintf("%s is preferred", pick.Name)
			alternates = append(alternates, ds)
		}
	}
	glog.Infof("Picked: %+v", pick)
	glog.Infof("Alternatives: %+v", alternates)
	glog.Infof("Rejects: %+v", rejects)
	return pick, alternates, rejects
}

// Status returns the status of a driver
func Status(name string) registry.DriverState {
	d := registry.Driver(name)
	return registry.DriverState{
		Name:     d.Name,
		Priority: d.Priority,
		State:    registry.Status(name),
	}
}

// SetLibvirtURI sets the URI to perform libvirt health checks against
func SetLibvirtURI(v string) {
	glog.Infof("Setting default libvirt URI to %s", v)
	os.Setenv("LIBVIRT_DEFAULT_URI", v)

}

// MachineName returns the name of the machine, as seen by the hypervisor given the cluster and node names
func MachineName(cc config.ClusterConfig, n config.Node) string {
	// For single node cluster, default to back to old naming
	if len(cc.Nodes) == 1 || n.ControlPlane {
		return cc.Name
	}
	return fmt.Sprintf("%s-%s", cc.Name, n.Name)
}
