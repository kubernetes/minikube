//go:build linux

/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package detect

import (
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sys/unix"
)

// cgroupVersion returns cgroup version as set on the linux OS host machine (where minikube runs).
// Possible options are: "v1", "v2" or "" (unknown).
// ref: https://kubernetes.io/docs/concepts/architecture/cgroups/#check-cgroup-version
// ref: https://man7.org/linux/man-pages/man7/cgroups.7.html
func cgroupVersion() string {
	if runtime.GOOS != "linux" {
		return ""
	}

	// check '/sys/fs/cgroup' or '/sys/fs/cgroup/unified' type
	var stat unix.Statfs_t
	if err := unix.Statfs("/sys/fs/cgroup", &stat); err != nil {
		return ""
	}
	// fallback, but could be misleading
	if stat.Type != unix.TMPFS_MAGIC && stat.Type != unix.CGROUP_SUPER_MAGIC && stat.Type != unix.CGROUP2_SUPER_MAGIC {
		if err := unix.Statfs("/sys/fs/cgroup/unified", &stat); err != nil {
			return ""
		}
	}

	switch stat.Type {
	case unix.TMPFS_MAGIC, unix.CGROUP_SUPER_MAGIC: // tmpfs, cgroupfs
		return "v1"
	case unix.CGROUP2_SUPER_MAGIC: // cgroup2fs
		return "v2"
	default:
		return ""
	}
}

func IsNinePSupported() bool {
	// assume true from non-linux
	if runtime.GOOS != "linux" {
		return true
	}
	_, err := os.Stat(getModuleRoot() + "/kernel/fs/9p")
	return err == nil
}

func getModuleRoot() string {
	// assume true from non-linux
	if runtime.GOOS != "linux" {
		return ""
	}
	uname := unix.Utsname{}
	if err := unix.Uname(&uname); err != nil {
		return ""
	}

	i := 0
	for ; uname.Release[i] != 0; i++ {
		continue
	}
	return filepath.Join("/lib/modules", string(uname.Release[:i]))

}
