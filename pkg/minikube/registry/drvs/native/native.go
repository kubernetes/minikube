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

package native

import (
	"fmt"
	"os/exec"
	"os/user"

	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/drivers/native"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.Native,
		Alias:    []string{driver.AliasNone},
		Config:   configure,
		Init:     func() drivers.Driver { return native.NewDriver(native.Config{}) },
		Status:   status,
		Priority: registry.Discouraged, // requires root
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	return native.NewDriver(native.Config{
		MachineName:      config.MachineName(cc, n),
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

	u, err := user.Current()
	if err != nil {
		return registry.State{Running: true, Error: err, Healthy: false, Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/none/"}
	}

	if u.Uid != "0" {
		test := exec.Command("sudo", "-n", "echo", "-n")
		if err := test.Run(); err != nil {
			return registry.State{Error: fmt.Errorf("running the 'none' driver as a regular user requires sudo permissions"), Healthy: false}
		}
	}
	return registry.State{Installed: true, Healthy: true}
}
