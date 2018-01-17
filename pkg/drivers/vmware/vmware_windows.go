/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package vmware

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var windowsInstallDir = `C:\Program Files (x86)\VMware\VMware Workstation`

func init() {
	// Parse HKEY_.CLASSES_ROOT\vm\shell\open\command's value like:
	// "C:\Program Files (x86)\VMware\VMware Workstation\vmware.exe" "%1"
	// in order to the Workstation install dir.
	key, err := registry.OpenKey(registry.CLASSES_ROOT, `vm\shell\open\command`, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	value, _, err := key.GetStringValue("")
	if err != nil {
		return
	}

	if value[0] == '"' {
		values := strings.Split(value[1:], "\"")
		windowsInstallDir = filepath.Dir(values[0])
	}
}

func DhcpConfigFiles() string {
	return `C:\ProgramData\VMware\vmnetdhcp.conf`
}

func DhcpLeaseFiles() string {
	return `C:\ProgramData\VMware\vmnetdhcp.leases`
}

func SetUmask() {
}

func setVmwareCmd(cmd string) string {
	cmd = cmd + ".exe"
	return filepath.Join(windowsInstallDir, cmd)
}

func getShareDriveAndName() (string, string, string) {
	return "Users", "C:\\Users", "/hosthome"
}
