/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"io/ioutil"
	"os/exec"

	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/style"
)

// HostInfo holds information on the user's machine
type HostInfo struct {
	Memory   int64
	CPUs     int
	DiskSize int64
}

func megs(bytes uint64) int64 {
	return int64(bytes / 1024 / 1024)
}

// CachedHostInfo returns system information such as memory,CPU, DiskSize
func CachedHostInfo() (*HostInfo, error, error, error) {
	var cpuErr, memErr, diskErr error
	i, cpuErr := cachedCPUInfo()
	if cpuErr != nil {
		glog.Warningf("Unable to get CPU info: %v", cpuErr)
	}
	v, memErr := cachedSysMemLimit()
	if memErr != nil {
		glog.Warningf("Unable to get mem info: %v", memErr)
	}

	d, diskErr := cachedDiskInfo()
	if diskErr != nil {
		glog.Warningf("Unable to get disk info: %v", diskErr)
	}

	var info HostInfo
	info.CPUs = len(i)
	info.Memory = megs(v.Total)
	info.DiskSize = megs(d.Total)
	return &info, cpuErr, memErr, diskErr
}

// showLocalOsRelease shows systemd information about the current linux distribution, on the local host
func showLocalOsRelease() {
	osReleaseOut, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		glog.Errorf("ReadFile: %v", err)
		return
	}

	osReleaseInfo, err := provision.NewOsRelease(osReleaseOut)
	if err != nil {
		glog.Errorf("NewOsRelease: %v", err)
		return
	}

	register.Reg.SetStep(register.LocalOSRelease)
	out.T(style.Provisioner, "OS release is {{.pretty_name}}", out.V{"pretty_name": osReleaseInfo.PrettyName})
}

// logRemoteOsRelease shows systemd information about the current linux distribution, on the remote VM
func logRemoteOsRelease(r command.Runner) {
	rr, err := r.RunCmd(exec.Command("cat", "/etc/os-release"))
	if err != nil {
		glog.Infof("remote release failed: %v", err)
	}

	osReleaseInfo, err := provision.NewOsRelease(rr.Stdout.Bytes())
	if err != nil {
		glog.Errorf("NewOsRelease: %v", err)
		return
	}

	glog.Infof("Remote host: %s", osReleaseInfo.PrettyName)
}

var (
	cachedSystemMemoryLimit *mem.VirtualMemoryStat
	cachedSystemMemoryErr   *error
)

//  cachedSysMemLimit will return a cached limit for the system's virtual memory.
func cachedSysMemLimit() (*mem.VirtualMemoryStat, error) {
	if cachedSystemMemoryLimit == nil {
		v, err := mem.VirtualMemory()
		cachedSystemMemoryLimit = v
		cachedSystemMemoryErr = &err
	}
	if cachedSystemMemoryErr == nil {
		return cachedSystemMemoryLimit, nil
	}
	return cachedSystemMemoryLimit, *cachedSystemMemoryErr
}

var (
	cachedDisk        *disk.UsageStat
	cachedDiskInfoErr *error
)

// cachedDiskInfo will return a cached disk usage info
func cachedDiskInfo() (disk.UsageStat, error) {
	if cachedDisk == nil {
		d, err := disk.Usage("/")
		cachedDisk = d
		cachedDiskInfoErr = &err
	}
	if cachedDiskInfoErr == nil {
		return *cachedDisk, nil
	}
	return *cachedDisk, *cachedDiskInfoErr
}

var (
	cachedCPU    *[]cpu.InfoStat
	cachedCPUErr *error
)

//  cachedCPUInfo will return a cached cpu info
func cachedCPUInfo() ([]cpu.InfoStat, error) {
	if cachedCPU == nil {
		i, err := cpu.Info()
		cachedCPU = &i
		cachedCPUErr = &err
	}
	if cachedCPUErr == nil {
		return *cachedCPU, nil
	}
	return *cachedCPU, *cachedCPUErr
}
