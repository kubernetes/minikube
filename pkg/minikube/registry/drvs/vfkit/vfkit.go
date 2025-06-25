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
	"crypto/rand"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/google/uuid"

	"k8s.io/minikube/pkg/drivers/vfkit"
	"k8s.io/minikube/pkg/drivers/vmnet"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
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
		Name:     driver.VFKit,
		Init:     func() drivers.Driver { return vfkit.NewDriver("", "") },
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

	switch cfg.Network {
	case "nat", "":
		// We generate a random mac address.
		var err error
		mac, err = generateMACAddress()
		if err != nil {
			return nil, fmt.Errorf("generating MAC address: %v", err)
		}
	case "vmnet-shared":
		// We generate a random UUID (or use a user provided one). vment-helper
		// will obtain a mac address from the vmnet framework using the UUID.
		u := cfg.UUID
		if u == "" {
			u = uuid.NewString()
		}
		helper = &vmnet.Helper{
			MachineDir:  filepath.Join(storePath, "machines", machineName),
			InterfaceID: u,
		}
	default:
		return nil, fmt.Errorf("unsupported network: %q", cfg.Network)
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
		Network:        cfg.Network,
		MACAddress:     mac,
		VmnetHelper:    helper,
	}, nil
}

func status() registry.State {
	_, err := exec.LookPath("vfkit")
	if err != nil {
		return registry.State{Error: err, Fix: "Run 'brew install vfkit'", Doc: docURL}
	}
	return registry.State{Installed: true, Healthy: true, Running: true}
}

func generateMACAddress() (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// Set local bit, ensure unicast address
	buf[0] = (buf[0] | 2) & 0xfe
	mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	return mac, nil
}
