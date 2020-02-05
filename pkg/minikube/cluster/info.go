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

package cluster

import (
	"io/ioutil"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
)

type hostInfo struct {
	Memory   int
	CPUs     int
	DiskSize int
}

func megs(bytes uint64) int {
	return int(bytes / 1024 / 1024)
}

func getHostInfo() (*hostInfo, error) {
	i, err := cpu.Info()
	if err != nil {
		glog.Warningf("Unable to get CPU info: %v", err)
		return nil, err
	}
	v, err := mem.VirtualMemory()
	if err != nil {
		glog.Warningf("Unable to get mem info: %v", err)
		return nil, err
	}
	d, err := disk.Usage("/")
	if err != nil {
		glog.Warningf("Unable to get disk info: %v", err)
		return nil, err
	}

	var info hostInfo
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

	out.T(out.Provisioner, "OS release is {{.pretty_name}}", out.V{"pretty_name": osReleaseInfo.PrettyName})
}

// logRemoteOsRelease shows systemd information about the current linux distribution, on the remote VM
func logRemoteOsRelease(drv drivers.Driver) {
	provisioner, err := provision.DetectProvisioner(drv)
	if err != nil {
		glog.Errorf("DetectProvisioner: %v", err)
		return
	}

	osReleaseInfo, err := provisioner.GetOsReleaseInfo()
	if err != nil {
		glog.Errorf("GetOsReleaseInfo: %v", err)
		return
	}

	glog.Infof("Provisioned with %s", osReleaseInfo.PrettyName)
}

// showHostInfo shows host information
func showHostInfo(cfg config.MachineConfig) {
	if driver.BareMetal(cfg.VMDriver) {
		info, err := getHostInfo()
		if err == nil {
			out.T(out.StartingNone, "Running on localhost (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"number_of_cpus": info.CPUs, "memory_size": info.Memory, "disk_size": info.DiskSize})
		}
		return
	}
	if driver.IsKIC(cfg.VMDriver) {
		info, err := getHostInfo() // TODO medyagh: get docker-machine info for non linux
		if err == nil {
			out.T(out.StartingVM, "Creating Kubernetes in {{.driver_name}} container with (CPUs={{.number_of_cpus}}), Memory={{.memory_size}}MB ({{.host_memory_size}}MB available) ...", out.V{"driver_name": cfg.VMDriver, "number_of_cpus": cfg.CPUs, "number_of_host_cpus": info.CPUs, "memory_size": cfg.Memory, "host_memory_size": info.Memory})
		}
		return
	}
	out.T(out.StartingVM, "Creating {{.driver_name}} VM (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"driver_name": cfg.VMDriver, "number_of_cpus": cfg.CPUs, "memory_size": cfg.Memory, "disk_size": cfg.DiskSize})
}
