// +build linux

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

package qemu

import (
	"fmt"
	"path/filepath"

	"github.com/docker/machine/libmachine/drivers"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	registry.Register(registry.DriverDef{
		Name:          "qemu",
		Builtin:       false,
		ConfigCreator: createQemuHost,
	})
}

type qemuDriver struct {
	*drivers.BaseDriver

	Memory         int
	DiskSize       int
	CPU            int
	Program        string
	Network        string
	SSHPort        int
	EnginePort     int
	ISO            string
	Boot2DockerURL string
	DiskPath       string
}

func createQemuHost(config cfg.MachineConfig) interface{} {
	return &qemuDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: cfg.GetMachineName(),
			StorePath:   constants.GetMinipath(),
			SSHUser:     "docker",
		},
		Memory:         config.Memory,
		CPU:            config.CPUs,
		Boot2DockerURL: config.Downloader.GetISOFileURI(config.MinikubeISO),
		DiskSize:       config.DiskSize,
		Program:        "qemu-system-x86_64",
		Network:        "user",
		SSHPort:        22,
		EnginePort:     2376,
		DiskPath:       filepath.Join(constants.GetMinipath(), "machines", cfg.GetMachineName(), fmt.Sprintf("%s.qcow2", cfg.GetMachineName())),
		ISO:            filepath.Join(constants.GetMinipath(), "machines", cfg.GetMachineName(), "boot2docker.iso"),
	}
}
