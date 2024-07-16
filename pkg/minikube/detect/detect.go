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

package detect

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/klauspost/cpuid"
	"github.com/spf13/viper"
	"golang.org/x/sys/cpu"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// RuntimeOS returns the runtime operating system
func RuntimeOS() string {
	return runtime.GOOS
}

// RuntimeArch returns the runtime architecture
func RuntimeArch() string {
	arch := runtime.GOARCH
	if arch == "arm" {
		// runtime.GOARM
		if !cpu.ARM.HasVFP {
			return "arm/v5"
		}
		if !cpu.ARM.HasVFPv3 {
			return "arm/v6"
		}
		// "arm" (== "arm/v7")
	}
	return arch
}

// IsMicrosoftWSL will return true if process is running in WSL in windows
// checking for WSL env var based on this https://github.com/microsoft/WSL/issues/423#issuecomment-608237689
// also based on https://github.com/microsoft/vscode/blob/90a39ba0d49d75e9a4d7e62a6121ad946ecebc58/resources/win32/bin/code.sh#L24
func IsMicrosoftWSL() bool {
	return os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSLPATH") != ""
}

// IsOnGCE determines whether minikube is currently running on GCE.
func IsOnGCE() bool {
	resp, err := http.Get("http://metadata.google.internal")
	if err != nil {
		return false
	}

	return resp.Header.Get("Metadata-Flavor") == "Google"
}

// IsCloudShell determines whether minikube is running inside CloudShell
func IsCloudShell() bool {
	e := os.Getenv("CLOUD_SHELL")
	return e == "true"
}

// IsAmd64M1Emulation  determines whether amd64 minikube binary is running on M1 mac in emulation mode
func IsAmd64M1Emulation() bool {
	return runtime.GOARCH == "amd64" && strings.HasPrefix(cpuid.CPU.BrandName, "VirtualApple")
}

// MinikubeInstalledViaSnap returns true if the minikube binary path includes "snap".
func MinikubeInstalledViaSnap() bool {
	ex, err := os.Executable()
	if err != nil {
		return false
	}
	exPath := filepath.Dir(ex)

	return strings.Contains(exPath, "snap")
}

// DockerInstalledViaSnap returns true if the Docker binary path includes "snap".
func DockerInstalledViaSnap() bool {
	c := exec.Command("which", "docker")
	o, err := c.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(o), "snap")
}

// GithubActionRunner returns true if running inside a github action runner
func GithubActionRunner() bool {
	// based on https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

// ImageCacheDir returns the path in the minikube home directory to the container image cache for the current architecture
func ImageCacheDir() string {
	return filepath.Join(localpath.MakeMiniPath("cache", "images"), runtime.GOARCH)
}

// KICCacheDir returns the path in the minikube home directory to the container node cache for the current architecture
func KICCacheDir() string {
	return filepath.Join(localpath.MakeMiniPath("cache", "kic"), runtime.GOARCH)
}

// ISOCacheDir returns the path in the minikube home directory to the virtual machine image cache for the current architecture
func ISOCacheDir() string {
	return filepath.Join(localpath.MakeMiniPath("cache", "iso"), runtime.GOARCH)
}

// SocketVMNetInstalled returns if socket_vmnet is installed
func SocketVMNetInstalled() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	return SocketVMNetPath() != "" && SocketVMNetClientPath() != ""
}

// SocketVMNetPath returns the path of socket_vmnet (QEMU driver only)
func SocketVMNetPath() string {
	p := viper.GetString("socket-vmnet-path")
	if p != "" {
		return p
	}
	return checkSocketVMNetInstallLocations("/var/run/socket_vmnet")
}

// SocketVMNetClientPath returns the path of socket_vmnet_client (QEMU driver only)
func SocketVMNetClientPath() string {
	p := viper.GetString("socket-vmnet-client-path")
	if p != "" {
		return p
	}
	return checkSocketVMNetInstallLocations("/opt/socket_vmnet/bin/socket_vmnet_client")
}

// checkSocketVMNetInstallLocations accepts a relative file path
// checks the three possible socket_vmnet install locations for existence of the file path
// if the file path exists it returns the full path, otherwise if returns an empty string
func checkSocketVMNetInstallLocations(path string) string {
	// source install, arm64 brew install, amd64 brew install
	prefixes := []string{"", "/opt/homebrew", "/usr/local"}
	for _, prefix := range prefixes {
		fullPath := prefix + path
		if fileExists(fullPath) {
			return fullPath
		}
	}
	return ""
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if !errors.Is(err, os.ErrNotExist) {
		klog.Warningf("failed to check for existence of %s: %v", filePath, err)
	}
	return false
}

// CgroupDriver returns detected cgroup driver as configured on host os.
// If unable to detect, it will return constants.DefaultCgroupDriver instead.
// ref: https://kubernetes.io/docs/setup/production-environment/container-runtimes/#cgroup-drivers
func CgroupDriver() string {
	switch cgroupVersion() {
	case "v1":
		klog.Infof("detected %q cgroup driver on host os", constants.CgroupfsCgroupDriver)
		return constants.CgroupfsCgroupDriver
	case "v2":
		klog.Infof("detected %q cgroup driver on host os", constants.SystemdCgroupDriver)
		return constants.SystemdCgroupDriver
	default:
		klog.Warningf("unable to detect host's os cgroup driver - will continue and try with %q as per default, but things might break", constants.DefaultCgroupDriver)
		return constants.DefaultCgroupDriver // try with default rather than just give up
	}
}

// MacOS13Plus returns if the current machine is running macOS 13+
func MacOS13Plus() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	o, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		klog.Warningf("failed to get macOS version: %v", err)
		return false
	}
	major, err := strconv.Atoi(strings.Split(string(o), ".")[0])
	if err != nil {
		klog.Warningf("failed to convert macOS version to int: %v", err)
		return false
	}
	return major >= 13
}
