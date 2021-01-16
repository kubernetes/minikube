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
	"errors"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/provision"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
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
	memory, err := parseMemFree(free)
	if err != nil {
		klog.Warningf("Unable to parse mem info: %v", err)
	}
	rr, diskErr := r.RunCmd(exec.Command("df", "-m"))
	if diskErr != nil {
		klog.Warningf("Unable to get disk info: %v", diskErr)
	}
	df := rr.Stdout.String()
	disksize, err := parseDiskFree(df)
	if err != nil {
		klog.Warningf("Unable to parse disk info: %v", err)
	}

	var info HostInfo
	info.CPUs = ncpus
	info.Memory = memory
	info.DiskSize = disksize
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

// ParseMemFree parses the output of the `free -m` command
func parseMemFree(out string) (int64, error) {
	//             total        used        free      shared  buff/cache   available
	//Mem:           1987         706         194           1        1086        1173
	//Swap:             0           0           0
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		if len(parsedLine) < 7 {
			continue
		}
		t, err := strconv.ParseInt(parsedLine[1], 10, 64)
		if err != nil {
			return 0, err
		}
		m := strings.Trim(parsedLine[0], ":")
		if m == "Mem" {
			return t, nil
		}
	}
	return 0, errors.New("no matching data found")
}

// ParseDiskFree parses the output of the `df -m` command
func parseDiskFree(out string) (int64, error) {
	// Filesystem     1M-blocks  Used Available Use% Mounted on
	// /dev/sda1          39643  3705     35922  10% /
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		if len(parsedLine) < 6 {
			continue
		}
		t, err := strconv.ParseInt(parsedLine[1], 10, 64)
		if err != nil {
			return 0, err
		}
		m := parsedLine[5]
		if m == "/" {
			return t, nil
		}
	}
	return 0, errors.New("no matching data found")
}
