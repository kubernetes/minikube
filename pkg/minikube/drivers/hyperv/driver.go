// +build windows

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

package hyperv

import (
	"github.com/docker/machine/drivers/hyperv"
	"github.com/docker/machine/libmachine/drivers"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	_ = registry.Register(registry.DriverDef{
		Name:          constants.DriverHyperv,
		Builtin:       true,
		ConfigCreator: createHypervHost,
		DriverCreator: func() drivers.Driver {
			return hyperv.NewDriver("", "")
		},
	})
}

func createHypervHost(config cfg.MachineConfig) interface{} {
	d := hyperv.NewDriver(cfg.GetMachineName(), localpath.MiniPath())

	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.VSwitch = config.HypervVirtualSwitch
	d.MemSize = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = config.DiskSize
	d.SSHUser = "docker"
	d.DisableDynamicMemory = true // default to disable dynamic memory as minikube is unlikely to work properly with dynamic memory

	return d
}
