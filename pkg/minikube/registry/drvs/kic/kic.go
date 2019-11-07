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
	"os/exec"

	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.Kic,
		Config:   configure,
		Init:     func() drivers.Driver { return kic.NewDriver(kic.Config{}) },
		Status:   status,
		Priority: registry.Discouraged, // requires root
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(mc config.MachineConfig) interface{} {
	return kic.NewDriver(kic.Config{
		MachineName:      config.GetMachineName(),
		StorePath:        localpath.MiniPath(),
		ContainerRuntime: mc.ContainerRuntime,
	})
}

func status() registry.State {
	_, err := exec.LookPath("docker")
	if err != nil {
		return registry.State{Error: err, Fix: "You need docker to use kic driver.", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/kic/"}
	}
	return registry.State{Installed: true, Healthy: true}
}
