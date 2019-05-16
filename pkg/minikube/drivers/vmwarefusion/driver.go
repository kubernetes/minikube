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

package vmwarefusion

import (
	"fmt"

	"github.com/docker/machine/drivers/vmwarefusion"
	"github.com/docker/machine/libmachine/drivers"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:          "vmwarefusion",
		Builtin:       true,
		ConfigCreator: createVMwareFusionHost,
		DriverCreator: func() drivers.Driver {
			return vmwarefusion.NewDriver("", "")
		},
	}); err != nil {
		panic(fmt.Sprintf("register: %v", err))
	}
}

func createVMwareFusionHost(config cfg.MachineConfig) interface{} {
	d := vmwarefusion.NewDriver(cfg.GetMachineName(), constants.GetMinipath()).(*vmwarefusion.Driver)
	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = config.DiskSize

	// TODO(philips): push these defaults upstream to fixup this driver
	d.SSHPort = 22
	d.ISO = d.ResolveStorePath("boot2docker.iso")
	return d
}
