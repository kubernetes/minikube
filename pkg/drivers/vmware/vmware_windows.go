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

	var (
		windowsInstallDir string
		err               error
	)
	for _, p := range []string{`SOFTWARE\WOW6432Node\VMware, Inc.`, `SOFTWARE\VMware, Inc.`} {
		windowsInstallDir, err = getVmwareInstallDirWithProductRegistryKey(p)
		if err == nil {
			return filepath.Join(windowsInstallDir, cmd)
		}
	}

	// panic in order to  allow to debug this issue more easily in the future
	panic(err)
}

// productRegistryKey is the registry entrypoint for the product
func getVmwareInstallDirWithProductRegistryKey(productRegistryKey string) (installDir string, err error) {

	rootPathKey, err := registry.OpenKey(registry.LOCAL_MACHINE, productRegistryKey, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer rootPathKey.Close()

	const DefaultVMwareCorePathKey = "Core"
	production, _, err := rootPathKey.GetStringValue(DefaultVMwareCorePathKey)
	if err != nil {
		return "", err
	}

	DefaultVMwareWSRegistryKey := productRegistryKey + `\` + production
	prodKey, err := registry.OpenKey(registry.LOCAL_MACHINE, DefaultVMwareWSRegistryKey, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer prodKey.Close()

	const DefaultVMwareWSInstallPathKey = "InstallPath"
	windowsInstallDir, _, err := prodKey.GetStringValue(DefaultVMwareWSInstallPathKey)
	if err != nil {
		return "", err
	}

	return windowsInstallDir, nil
}
