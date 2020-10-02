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

package none

import (
	"fmt"
	"os/exec"
	"os/user"

	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/drivers/none"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.None,
		Config:   configure,
		Init:     func() drivers.Driver { return none.NewDriver(none.Config{}) },
		Status:   status,
		Priority: registry.Discouraged, // requires root
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	return none.NewDriver(none.Config{
		MachineName:      driver.MachineName(cc, n),
		StorePath:        localpath.MiniPath(),
		ContainerRuntime: cc.KubernetesConfig.ContainerRuntime,
	}), nil
}

func status() registry.State {
	_, err := exec.LookPath("iptables")
	if err != nil {
		return registry.State{Running: true, Error: err, Fix: "iptables must be installed", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/none/"}
	}

	if _, err := exec.LookPath("docker"); err != nil {
		return registry.State{Running: true, Error: err, Installed: false, Fix: "Install docker", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/none/"}
	}

	_, err = user.Current()
	if err != nil {
		return registry.State{Running: true, Error: err, Healthy: false, Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/none/"}
	}

	return registry.State{Installed: true, Healthy: true}
}
