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

	"github.com/docker/machine/libmachine/drivers"

	"k8s.io/minikube/pkg/drivers/vfkit"
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
	if err := registry.Register(registry.DriverDef{
		Name:     driver.VFKit,
		Init:     func() drivers.Driver { return vfkit.NewDriver("", "") },
		Config:   configure,
		Status:   status,
		Default:  true,
		Priority: registry.Experimental,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cfg config.ClusterConfig, n config.Node) (interface{}, error) {
	mac, err := generateMACAddress()
	if err != nil {
		return nil, fmt.Errorf("generating MAC address: %v", err)
	}

	return &vfkit.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: config.MachineName(cfg, n),
			StorePath:   localpath.MiniPath(),
			SSHUser:     "docker",
		},
		Boot2DockerURL: download.LocalISOResource(cfg.MinikubeISO),
		DiskSize:       cfg.DiskSize,
		Memory:         cfg.Memory,
		CPU:            cfg.CPUs,
		MACAddress:     mac,
		Cmdline:        "",
		ExtraDisks:     cfg.ExtraDisks,
	}, nil
}

func status() registry.State {
	_, err := exec.LookPath("vfkit")
	if err != nil {
		return registry.State{Error: err, Fix: "Run 'brew tap cfergeau/crc && brew install vfkit'", Doc: docURL}
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
