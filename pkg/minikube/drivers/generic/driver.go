/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package generic

import (
	"fmt"

	"github.com/docker/machine/drivers/generic"
	"github.com/docker/machine/libmachine/drivers"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	err := registry.Register(registry.DriverDef{
		Name:          constants.DriverGeneric,
		Builtin:       true,
		ConfigCreator: createGenericHost,
		DriverCreator: func() drivers.Driver {
			return generic.NewDriver("", "")
		},
	})
	if err != nil {
		panic(fmt.Sprintf("unable to register: %v", err))
	}
}

func createGenericHost(config cfg.MachineConfig) interface{} {
	d := generic.NewDriver(cfg.GetMachineName(), constants.GetMinipath())

	if config.GenericIPAddress == "" {
		exit.UsageT(`Please provide an IP address. the vm-driver "{{.driver_name}}" requires it.`, out.V{"driver_name": constants.DriverGeneric})
	}

	d.(*generic.Driver).IPAddress = config.GenericIPAddress
	d.(*generic.Driver).SSHUser = config.GenericSSHUser
	d.(*generic.Driver).SSHKey = config.GenericSSHKey
	d.(*generic.Driver).SSHPort = config.GenericSSHPort

	return d
}
