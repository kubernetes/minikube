// +build darwin

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

package hyperkit

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/pborman/uuid"
	"k8s.io/minikube/pkg/drivers/hyperkit"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:          "hyperkit",
		Builtin:       false,
		ConfigCreator: createHyperkitHost,
	}); err != nil {
		panic(fmt.Sprintf("register: %v", err))
	}
}

func createHyperkitHost(config cfg.MachineConfig) interface{} {
	uuID := config.UUID
	if uuID == "" {
		uuID = uuid.NewUUID().String()
	}

	return &hyperkit.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: cfg.GetMachineName(),
			StorePath:   constants.GetMinipath(),
			SSHUser:     "docker",
		},
		Boot2DockerURL: config.Downloader.GetISOFileURI(config.MinikubeISO),
		DiskSize:       config.DiskSize,
		Memory:         config.Memory,
		CPU:            config.CPUs,
		NFSShares:      config.NFSShare,
		NFSSharesRoot:  config.NFSSharesRoot,
		UUID:           uuID,
		VpnKitSock:     config.HyperkitVpnKitSock,
		VSockPorts:     config.HyperkitVSockPorts,
		Cmdline:        "loglevel=3 user=docker console=ttyS0 console=tty0 noembed nomodeset norestore waitusb=10 systemd.legacy_systemd_cgroup_controller=yes base host=" + cfg.GetMachineName(),
	}
}
