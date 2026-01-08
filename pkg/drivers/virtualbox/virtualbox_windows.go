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

package virtualbox

import (
	"strings"

	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
	"k8s.io/minikube/pkg/libmachine/log"
)

// cmdOutput runs a shell command and returns its output.
func cmdOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	log.Debugf("COMMAND: %v %v", name, strings.Join(args, " "))

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	log.Debugf("STDOUT:\n{\n%v}", string(stdout))

	return string(stdout), nil
}

func detectVBoxManageCmd() string {
	cmd := "VBoxManage"
	if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}

	if p := os.Getenv("VBOX_MSI_INSTALL_PATH"); p != "" {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}

	// Look in default installation path for VirtualBox version > 5
	if path, err := exec.LookPath(filepath.Join("C:\\Program Files\\Oracle\\VirtualBox", cmd)); err == nil {
		return path
	}

	// Look in windows registry
	if p, err := findVBoxInstallDirInRegistry(); err == nil {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}

	return detectVBoxManageCmdInPath() //fallback to path
}

func findVBoxInstallDirInRegistry() (string, error) {
	registryKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Oracle\VirtualBox`, registry.QUERY_VALUE)
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find VirtualBox registry entries, is VirtualBox really installed properly? %s", err)
		log.Debugf(errorMessage)
		return "", fmt.Errorf("%s", errorMessage)
	}

	defer registryKey.Close()

	installDir, _, err := registryKey.GetStringValue("InstallDir")
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find InstallDir registry key within VirtualBox registries entries, is VirtualBox really installed properly? %s", err)
		log.Debugf(errorMessage)
		return "", fmt.Errorf("%s", errorMessage)
	}

	return installDir, nil
}

func getShareDriveAndName() (string, string) {
	return "c/Users", "\\\\?\\c:\\Users"
}

func isHyperVInstalled() bool {
	// check if hyper-v is installed
	_, err := exec.LookPath("vmms.exe")
	if err != nil {
		errmsg := "Hyper-V is not installed."
		log.Debugf(errmsg, err)
		return false
	}

	// check to see if a hypervisor is present. if hyper-v is installed and enabled,
	// display an error explaining the incompatibility between virtualbox and hyper-v.
	output, err := cmdOutput("wmic", "computersystem", "get", "hypervisorpresent")

	if err != nil {
		errmsg := "Could not check to see if Hyper-V is running."
		log.Debugf(errmsg, err)
		return false
	}

	enabled := strings.Contains(output, "TRUE")
	return enabled

}
