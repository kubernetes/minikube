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
func CachedHostInfo() (*HostInfo, error) {
	i, err := cachedCPUInfo()
	if err != nil {
		glog.Warningf("Unable to get CPU info: %v", err)
		return nil, err
	}
	v, err := cachedSysMemLimit()
	if err != nil {
		glog.Warningf("Unable to get mem info: %v", err)
		return nil, err
	}

	d, err := cachedDisInfo()
	if err != nil {
		glog.Warningf("Unable to get disk info: %v", err)
		return nil, err
	}

	var info HostInfo
	info.CPUs = len(i)
	info.Memory = megs(v.Total)
	info.DiskSize = megs(d.Total)
	return &info, nil
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
	out.T(out.Provisioner, "OS release is {{.pretty_name}}", out.V{"pretty_name": osReleaseInfo.PrettyName})
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

var cachedSystemMemoryLimit *mem.VirtualMemoryStat
var cachedSystemMemoryErr *error

//  cachedSysMemLimit will return a cached limit for the system's virtual memory.
func cachedSysMemLimit() (*mem.VirtualMemoryStat, error) {
	if cachedSystemMemoryLimit == nil {
		v, err := mem.VirtualMemory()
		cachedSystemMemoryLimit = v
		cachedSystemMemoryErr = &err
	}
	return cachedSystemMemoryLimit, *cachedSystemMemoryErr
}

var cachedDiskInfo *disk.UsageStat
var cachedDiskInfoeErr *error

// cachedDisInfo will return a cached disk usage info
func cachedDisInfo() (disk.UsageStat, error) {
	if cachedDiskInfo == nil {
		d, err := disk.Usage("/")
		cachedDiskInfo = d
		cachedDiskInfoeErr = &err
	}
	return *cachedDiskInfo, *cachedDiskInfoeErr
}

var cachedCPU *[]cpu.InfoStat
var cachedCPUErr *error

//  cachedCPUInfo will return a cached cpu info
func cachedCPUInfo() ([]cpu.InfoStat, error) {
	if cachedCPU == nil {
		i, err := cpu.Info()
		cachedCPU = &i
		cachedCPUErr = &err
		if err != nil {
			return nil, *cachedCPUErr
		}
	}
	return *cachedCPU, *cachedCPUErr
}
