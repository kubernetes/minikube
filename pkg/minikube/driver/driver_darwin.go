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
	"os/exec"
	"runtime"
	"strings"
)

// supportedDrivers is a list of supported drivers on Darwin.
var supportedDrivers = func() []string {
	if runtime.GOARCH == "arm64" {
		// on darwin/arm64 only docker and ssh are supported yet
		return []string{
			QEMU2,
			VFKit,
			Krunkit,
			Parallels,
			Docker,
			Podman,
			SSH,
		}
	}
	// PowerPC does not support podman
	if strings.HasPrefix(runtime.GOARCH, "ppc") {
		return []string{
			VirtualBox,
			Parallels,
			HyperKit,
			VMware,
			Docker,
			SSH,
		}
	}
	return []string{
		VirtualBox,
		Parallels,
		HyperKit,
		VMware,
		QEMU2,
		VFKit,
		Docker,
		Podman,
		SSH,
	}
}()

func VBoxManagePath() string {
	cmd := "VBoxManage"
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}
