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

package parallels

import (
	"fmt"

	parallels "github.com/Parallels/docker-machine-parallels"
	"github.com/docker/machine/libmachine/drivers"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	err := registry.Register(registry.DriverDef{
		Name:          "parallels",
		Builtin:       true,
		ConfigCreator: createParallelsHost,
		DriverCreator: func() drivers.Driver {
			return parallels.NewDriver("", "")
		},
	})
	if err != nil {
		panic(fmt.Sprintf("unable to register: %v", err))
	}

}

func createParallelsHost(config cfg.MachineConfig) interface{} {
	d := parallels.NewDriver(cfg.GetMachineName(), constants.GetMinipath()).(*parallels.Driver)
	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = config.DiskSize
	return d
}
