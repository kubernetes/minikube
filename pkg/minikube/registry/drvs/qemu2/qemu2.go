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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/common/mac"
	"k8s.io/minikube/pkg/drivers/qemu"
	"k8s.io/minikube/pkg/libmachine/drivers"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/run"
)

const docURL = "https://minikube.sigs.k8s.io/docs/reference/drivers/qemu/"

func init() {
	priority := registry.Default
	if runtime.GOOS == "windows" {
		priority = registry.Experimental
	}
	if err := registry.Register(registry.DriverDef{
		Name:  driver.QEMU2,
		Alias: []string{driver.AliasQEMU},
		Init: func(options *run.CommandOptions) drivers.Driver {
			return qemu.NewDriver("", "", options)
		},
		Config:       configure,
		Status:       status,
		Default:      true,
		Priority:     priority,
		Parallel:     true,
		ProbeTimeout: 1 * time.Second,
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
		return semver.Version{}, fmt.Errorf("%s -version: %w", qemuSystem, err)
	}
	return parseQemuVersion(string(rr))
}

// qemuVersionRegexp matches the version reported by `qemu-system-* -version`.
// Distributions and custom builds append arbitrary suffixes after the version
// number, so only the leading "major.minor[.patch]" is captured, e.g.:
//
//	QEMU emulator version 8.2.1 (v8.2.1)
//	QEMU emulator version 7.2.0 (Debian 1:7.2+dfsg-7)
//	QEMU emulator version 2.11.1(Debian 1:2.11+dfsg-1ubuntu7.42)
var qemuVersionRegexp = regexp.MustCompile(`QEMU emulator version (\d+)\.(\d+)(?:\.(\d+))?`)

// parseQemuVersion extracts the semantic version from the output of
// `qemu-system-* -version`. A missing patch component is treated as 0.
func parseQemuVersion(output string) (semver.Version, error) {
	m := qemuVersionRegexp.FindStringSubmatch(output)
	if m == nil {
		firstLine, _, _ := strings.Cut(strings.TrimSpace(output), "\n")
		return semver.Version{}, fmt.Errorf("unable to parse QEMU version from %q", strings.TrimSpace(firstLine))
	}
	parts := [3]uint64{}
	for i, s := range m[1:] {
		if s == "" {
			continue
		}
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return semver.Version{}, fmt.Errorf("invalid QEMU version component %q: %w", s, err)
		}
		parts[i] = n
	}
	return semver.Version{Major: parts[0], Minor: parts[1], Patch: parts[2]}, nil
}

// needsHighmemOff reports whether the arm64 "virt" machine on darwin must be
// started with highmem=off. QEMU before 7.0.0 requires it, and highmem also
// fails with 3072MB of memory or less.
func needsHighmemOff(v semver.Version, memoryMB int) bool {
	return v.LT(semver.Version{Major: 7}) || memoryMB <= 3072
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
			v, err := qemuVersion()
			if err != nil {
				return nil, err
			}
			if needsHighmemOff(v, cc.Memory) {
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
	macAddr := mac.FromName(name)
	klog.Infof("Using mac address %s", macAddr)

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
		MACAddress:            macAddr,
		SocketVMNetPath:       cc.SocketVMnetPath,
		SocketVMNetClientPath: cc.SocketVMnetClientPath,
		ExtraDisks:            cc.ExtraDisks,
	}, nil
}

func status(_ *run.CommandOptions) registry.State {
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
