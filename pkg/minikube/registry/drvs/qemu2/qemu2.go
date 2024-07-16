/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package qemu2

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/drivers/qemu"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

const docURL = "https://minikube.sigs.k8s.io/docs/reference/drivers/qemu/"

func init() {
	priority := registry.Default
	if runtime.GOOS == "windows" {
		priority = registry.Experimental
	}
	if err := registry.Register(registry.DriverDef{
		Name:     driver.QEMU2,
		Alias:    []string{driver.AliasQEMU},
		Init:     func() drivers.Driver { return qemu.NewDriver("", "") },
		Config:   configure,
		Status:   status,
		Default:  true,
		Priority: priority,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func qemuSystemProgram() (string, error) {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		return "qemu-system-x86_64", nil
	case "arm64":
		return "qemu-system-aarch64", nil
	default:
		return "", fmt.Errorf("unknown arch: %s", arch)
	}
}

func qemuFirmwarePath(customPath string) (string, error) {
	if customPath != "" {
		return customPath, nil
	}
	if runtime.GOOS == "windows" {
		return "C:\\Program Files\\qemu\\share\\edk2-x86_64-code.fd", nil
	}
	if detect.IsAmd64M1Emulation() {
		return "/opt/homebrew/opt/qemu/share/qemu/edk2-x86_64-code.fd", nil
	}
	arch := runtime.GOARCH
	// For macOS, find the correct brew installation path for qemu firmware
	if runtime.GOOS == "darwin" {
		switch arch {
		case "amd64":
			return "/usr/local/opt/qemu/share/qemu/edk2-x86_64-code.fd", nil
		case "arm64":
			return "/opt/homebrew/opt/qemu/share/qemu/edk2-aarch64-code.fd", nil
		default:
			return "", fmt.Errorf("unknown arch: %s", arch)
		}
	}

	switch arch {
	case "amd64":
		return "/usr/share/OVMF/OVMF_CODE.fd", nil
	case "arm64":
		return "/usr/share/AAVMF/AAVMF_CODE.fd", nil
	default:
		return "", fmt.Errorf("unknown arch: %s", arch)
	}
}

func qemuVersion() (semver.Version, error) {
	qemuSystem, err := qemuSystemProgram()
	if err != nil {
		return semver.Version{}, err
	}

	cmd := exec.Command(qemuSystem, "-version")
	rr, err := cmd.Output()
	if err != nil {
		return semver.Version{}, err
	}
	v := strings.Split(strings.TrimPrefix(string(rr), "QEMU emulator version "), "\n")[0]
	return semver.Make(v)
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	name := config.MachineName(cc, n)
	qemuSystem, err := qemuSystemProgram()
	if err != nil {
		return nil, err
	}
	var qemuMachine string
	var qemuCPU string
	switch runtime.GOARCH {
	case "amd64":
		qemuMachine = "" // default
		// set cpu type to max to enable higher microarchitecture levels
		// see https://lists.gnu.org/archive/html/qemu-devel/2022-08/msg04066.html for details
		qemuCPU = "max"
	case "arm64":
		qemuMachine = "virt"
		qemuCPU = "cortex-a72"
		// highmem=off needed for qemu 6.2.0 and lower, see https://patchwork.kernel.org/project/qemu-devel/patch/20201126215017.41156-9-agraf@csgraf.de/#23800615 for details
		if runtime.GOOS == "darwin" {
			qemu7 := semver.MustParse("7.0.0")
			v, err := qemuVersion()
			if err != nil {
				return nil, err
			}
			// Surprisingly, highmem doesn't work for low memory situations
			if v.LT(qemu7) || cc.Memory <= 3072 {
				qemuMachine += ",highmem=off"
			}
			qemuCPU = "host"
		} else if _, err := os.Stat("/dev/kvm"); err == nil {
			qemuMachine += ",gic-version=3"
			qemuCPU = "host"
		}
	default:
		return nil, fmt.Errorf("unknown arch: %s", runtime.GOARCH)
	}
	qemuFirmware, err := qemuFirmwarePath(cc.CustomQemuFirmwarePath)
	if err != nil {
		return nil, err
	}
	mac, err := generateMACAddress()
	if err != nil {
		return nil, fmt.Errorf("generating MAC address: %v", err)
	}

	return qemu.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: name,
			StorePath:   localpath.MiniPath(),
			SSHUser:     "docker",
		},
		Boot2DockerURL:        download.LocalISOResource(cc.MinikubeISO),
		DiskSize:              cc.DiskSize,
		Memory:                cc.Memory,
		CPU:                   cc.CPUs,
		EnginePort:            2376,
		FirstQuery:            true,
		DiskPath:              filepath.Join(localpath.MiniPath(), "machines", name, fmt.Sprintf("%s.img", name)),
		Program:               qemuSystem,
		BIOS:                  runtime.GOARCH != "arm64",
		MachineType:           qemuMachine,
		CPUType:               qemuCPU,
		Firmware:              qemuFirmware,
		VirtioDrives:          false,
		Network:               cc.Network,
		CacheMode:             "default",
		IOMode:                "threads",
		MACAddress:            mac,
		SocketVMNetPath:       cc.SocketVMnetPath,
		SocketVMNetClientPath: cc.SocketVMnetClientPath,
		ExtraDisks:            cc.ExtraDisks,
	}, nil
}

func status() registry.State {
	qemuSystem, err := qemuSystemProgram()
	if err != nil {
		return registry.State{Error: err, Doc: docURL}
	}

	if _, err := exec.LookPath(qemuSystem); err != nil {
		return registry.State{Error: err, Fix: "Install qemu-system", Doc: docURL}
	}

	qemuFirmware, err := qemuFirmwarePath(viper.GetString("qemu-firmware-path"))
	if err != nil {
		return registry.State{Error: err, Doc: docURL}
	}

	if _, err := os.Stat(qemuFirmware); err != nil {
		return registry.State{Error: err, Fix: "Install uefi firmware", Doc: docURL}
	}

	return registry.State{Installed: true, Healthy: true, Running: true}
}

func generateMACAddress() (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// Set local bit, ensure unicast address, socket_vmnet doesn't support multicast
	buf[0] = (buf[0] | 2) & 0xfe
	mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	return mac, nil
}
