//go:build darwin

/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package vfkit

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"k8s.io/minikube/pkg/libmachine/drivers"

	"k8s.io/minikube/pkg/drivers/common"
	"k8s.io/minikube/pkg/drivers/common/virtiofs"
	"k8s.io/minikube/pkg/drivers/common/vmnet"
	"k8s.io/minikube/pkg/drivers/vfkit"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/run"
)

const (
	docURL = "https://minikube.sigs.k8s.io/docs/reference/drivers/vfkit/"
)

func init() {
	priority := registry.Unknown
	if runtime.GOOS == "darwin" {
		priority = registry.Preferred
	}
	if err := registry.Register(registry.DriverDef{
		Name: driver.VFKit,
		Init: func(options *run.CommandOptions) drivers.Driver {
			return vfkit.NewDriver("", "", options)
		},
		Config:   configure,
		Status:   status,
		Default:  true,
		Priority: priority,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cfg config.ClusterConfig, n config.Node) (interface{}, error) {
	var mac string
	var helper *vmnet.Helper

	machineName := config.MachineName(cfg, n)
	storePath := localpath.MiniPath()

	switch {
	case cfg.Network == "nat" || cfg.Network == "":
		// We generate a random mac address.
		var err error
		mac, err = common.GenerateMACAddress()
		if err != nil {
			return nil, fmt.Errorf("generating MAC address: %v", err)
		}
	case cfg.Network == "vmnet-shared" || strings.HasPrefix(cfg.Network, "vmnet:"):
		// Normalize legacy "vmnet-shared" to "vmnet:shared".
		networkName := cfg.Network
		if networkName == "vmnet-shared" {
			networkName = "vmnet:shared"
		}
		var u string
		if vmnet.GeneratesMACAddress() {
			// Interface-id mode: generate UUID, vmnet will provide MAC.
			u = cfg.UUID
			if u == "" {
				u = uuid.NewString()
			}
		} else {
			// Network mode: we generate the MAC address.
			var err error
			mac, err = common.GenerateMACAddress()
			if err != nil {
				return nil, fmt.Errorf("generating MAC address: %v", err)
			}
		}
		helper = &vmnet.Helper{
			MachineDir:  filepath.Join(storePath, "machines", machineName),
			InterfaceID: u,
			NetworkName: networkName,
		}
	default:
		return nil, fmt.Errorf("unsupported network: %q", cfg.Network)
	}

	mounts, err := virtiofs.ValidateMountString(cfg.MountString)
	if err != nil {
		return nil, err
	}

	return &vfkit.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: machineName,
			StorePath:   storePath,
			SSHUser:     "docker",
		},
		Boot2DockerURL: download.LocalISOResource(cfg.MinikubeISO),
		DiskSize:       cfg.DiskSize,
		Memory:         cfg.Memory,
		CPU:            cfg.CPUs,
		ExtraDisks:     cfg.ExtraDisks,
		VirtiofsMounts: mounts,
		Network:        cfg.Network,
		MACAddress:     mac,
		VmnetHelper:    helper,
		Rosetta:        cfg.Rosetta,
	}, nil
}

func status(_ *run.CommandOptions) registry.State {
	_, err := exec.LookPath("vfkit")
	if err != nil {
		return registry.State{Error: err, Fix: "Run 'brew install vfkit'", Doc: docURL}
	}
	return registry.State{Installed: true, Healthy: true, Running: true}
}
