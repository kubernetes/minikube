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

package none

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/drivers/none"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:          "none",
		Builtin:       true,
		ConfigCreator: createNoneHost,
		DriverCreator: func() drivers.Driver {
			return none.NewDriver(none.Config{})
		},
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

// createNoneHost creates a none Driver from a MachineConfig
func createNoneHost(config cfg.MachineConfig) interface{} {
	return none.NewDriver(none.Config{
		MachineName:      cfg.GetMachineName(),
		StorePath:        constants.GetMinipath(),
		ContainerRuntime: config.ContainerRuntime,
	})
}
