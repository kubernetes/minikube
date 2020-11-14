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
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/provision"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util"
)

// HostInfo holds information on the user's machine
type HostInfo struct {
	Memory   int64
	CPUs     int
	DiskSize int64
}

// LocalHostInfo returns system information such as memory,CPU, DiskSize
func LocalHostInfo() (*HostInfo, error, error, error) {
	var cpuErr, memErr, diskErr error
	i, cpuErr := cachedCPUInfo()
	if cpuErr != nil {
		klog.Warningf("Unable to get CPU info: %v", cpuErr)
	}
	v, memErr := cachedSysMemLimit()
	if memErr != nil {
		klog.Warningf("Unable to get mem info: %v", memErr)
	}

	d, diskErr := cachedDiskInfo()
	if diskErr != nil {
		klog.Warningf("Unable to get disk info: %v", diskErr)
	}

	var info HostInfo
	info.CPUs = len(i)
	info.Memory = util.ConvertUnsignedBytesToMB(v.Total)
	info.DiskSize = util.ConvertUnsignedBytesToMB(d.Total)
	return &info, cpuErr, memErr, diskErr
}

// RemoteHostInfo returns system information such as memory,CPU, DiskSize
func RemoteHostInfo(r command.Runner) (*HostInfo, error, error, error) {
	rr, cpuErr := r.RunCmd(exec.Command("nproc"))
	if cpuErr != nil {
		klog.Warningf("Unable to get CPU info: %v", cpuErr)
	}
	nproc := rr.Stdout.String()
	ncpus, err := strconv.Atoi(strings.TrimSpace(nproc))
	if err != nil {
		klog.Warningf("Failed to parse CPU info: %v", err)
	}
	rr, memErr := r.RunCmd(exec.Command("free", "-m"))
	if memErr != nil {
		klog.Warningf("Unable to get mem info: %v", memErr)
	}
	free := rr.Stdout.String()
	memory, _, err := util.ParseMemFree(free)
	if err != nil {
		klog.Warningf("Unable to parse mem info: %v", err)
	}
	rr, diskErr := r.RunCmd(exec.Command("df", "-m"))
	if diskErr != nil {
		klog.Warningf("Unable to get disk info: %v", diskErr)
	}
	df := rr.Stdout.String()
	disksize, _, err := util.ParseDiskFree(df, "/")
	if err != nil {
		klog.Warningf("Unable to parse disk info: %v", err)
	}

	var info HostInfo
	info.CPUs = ncpus
	info.Memory = int64(memory)
	info.DiskSize = int64(disksize)
	return &info, cpuErr, memErr, diskErr
}

// showLocalOsRelease shows systemd information about the current linux distribution, on the local host
func showLocalOsRelease() {
	osReleaseOut, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		klog.Errorf("ReadFile: %v", err)
		return
	}

	osReleaseInfo, err := provision.NewOsRelease(osReleaseOut)
	if err != nil {
		klog.Errorf("NewOsRelease: %v", err)
		return
	}

	register.Reg.SetStep(register.LocalOSRelease)
	out.Step(style.Provisioner, "OS release is {{.pretty_name}}", out.V{"pretty_name": osReleaseInfo.PrettyName})
}

// logRemoteOsRelease shows systemd information about the current linux distribution, on the remote VM
func logRemoteOsRelease(r command.Runner) {
	rr, err := r.RunCmd(exec.Command("cat", "/etc/os-release"))
	if err != nil {
		klog.Infof("remote release failed: %v", err)
	}

	osReleaseInfo, err := provision.NewOsRelease(rr.Stdout.Bytes())
	if err != nil {
		klog.Errorf("NewOsRelease: %v", err)
		return
	}

	klog.Infof("Remote host: %s", osReleaseInfo.PrettyName)
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
