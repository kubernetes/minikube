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

package kic

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/drivers/kic"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:          constants.DriverKic,
		Builtin:       true,
		ConfigCreator: createKicHost,
		DriverCreator: func() drivers.Driver {
			fmt.Println("Inside pkg/m/drivers/kic init")
			return kic.NewDriver(kic.Config{}) // MEDYA:TODO verify this gets filled
		},
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

// createKicHost creates a kick Driver from a MachineConfig
func createKicHost(config cfg.MachineConfig) interface{} {
	fmt.Println("Inside createKicHost in pkg/m/drivers/kic")
	return kic.NewDriver(kic.Config{
		MachineName:      cfg.GetMachineName(),
		StorePath:        constants.GetMinipath(), // MEDYA:TODO check what is stored here.
		ContainerRuntime: config.ContainerRuntime, // MEDYA:TODO might not need.
		OciClient:        config.OciClient,
		KicImage:         config.KicImage,
	})
}
