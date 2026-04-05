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
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// https://docs.microsoft.com/en-us/windows/deployment/usmt/usmt-recognized-environment-variables

func DhcpConfigFiles() string {
	return filepath.Join(os.Getenv("ALLUSERSPROFILE"), `VMware\vmnetdhcp.conf`)
}

func DhcpLeaseFiles() string {
	return filepath.Join(os.Getenv("ALLUSERSPROFILE"), `VMware\vmnetdhcp.leases`)
}

func SetUmask() {
}

func setVmwareCmd(cmd string) string {
	cmd = cmd + ".exe"
	DefaultVMWareWSProductionRegistryKey := `SOFTWARE\WOW6432Node\VMware, Inc.`
	DefaultVMwareCorePathKey := "Core"
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, DefaultVMWareWSProductionRegistryKey, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	production, _, err := k.GetStringValue(DefaultVMwareCorePathKey)
	if err != nil {
		return ""
	}

	//Get the VMware Product Install Path
	DefaultVMwareWSRegistryKey := fmt.Sprintf(`SOFTWARE\WOW6432Node\VMware, Inc.\%s`, production)
	DefaultVMwareWSInstallPathKey := "InstallPath"

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, DefaultVMwareWSRegistryKey, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()

	value, _, err := key.GetStringValue(DefaultVMwareWSInstallPathKey)
	if err != nil {
		return ""
	}
	windowsInstallDir := value
	return filepath.Join(windowsInstallDir, cmd)
}
