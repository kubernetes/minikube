//go:build linux

/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"fmt"
	"os"
	"path"

	"github.com/opencontainers/runc/libcontainer/cgroups"

	"k8s.io/klog/v2"
)

// findCgroupMountpoints returns the cgroups mount point
// defined in docker engine engine/pkg/sysinfo/sysinfo_linux.go
func findCgroupMountpoints() (map[string]string, error) {
	cgMounts, err := cgroups.GetCgroupMounts(false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cgroup information: %v", err)
	}
	mps := make(map[string]string)
	for _, m := range cgMounts {
		for _, ss := range m.Subsystems {
			mps[ss] = m.Mountpoint
		}
	}
	return mps, nil
}

// HasMemoryCgroup checks whether it is possible to set memory limit for cgroup.
func HasMemoryCgroup() bool {
	cgMounts, err := findCgroupMountpoints()
	if err != nil {
		klog.Warning("Your kernel does not support memory limit capabilities or the cgroup is not mounted.")
		return false
	}
	_, ok := cgMounts["memory"]
	if !ok {
		klog.Warning("Your kernel does not support memory limit capabilities or the cgroup is not mounted.")
		return false
	}
	return true
}

// hasMemorySwapCgroup checks whether it is possible to set swap limit for cgroup
func hasMemorySwapCgroup() bool {
	cgMounts, err := findCgroupMountpoints()
	if err != nil {
		klog.Warning("Your kernel does not support swap limit capabilities or the cgroup is not mounted.")
		return false
	}
	mountPoint, ok := cgMounts["memory"]
	if !ok {
		klog.Warning("Your kernel does not support swap limit capabilities or the cgroup is not mounted.")
		return false
	}

	_, err = os.Stat(path.Join(mountPoint, "memory.memsw.limit_in_bytesw"))
	if err != nil {
		klog.Warning("Your kernel does not support swap limit capabilities or the cgroup is not mounted.")
		return false

	}
	return true
}
