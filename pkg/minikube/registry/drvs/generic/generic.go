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

	"github.com/docker/machine/libmachine/drivers"
	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/drivers/generic"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	err := registry.Register(registry.DriverDef{
		Name:     driver.Generic,
		Config:   configure,
		Status:   status,
		Priority: registry.Fallback,
		Init:     func() drivers.Driver { return generic.NewDriver("", "") },
	})
	if err != nil {
		panic(fmt.Sprintf("unable to register: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	d := generic.NewDriver(driver.MachineName(cc, n), localpath.MiniPath())

	if cc.GenericIPAddress == "" {
		return nil, errors.Errorf("please provide an IP address")
	}

	d.IPAddress = cc.GenericIPAddress
	d.SSHUser = cc.GenericSSHUser
	d.SSHKey = cc.GenericSSHKey
	d.SSHPort = cc.GenericSSHPort

	return d, nil
}

func status() registry.State {
	return registry.State{Installed: true, Healthy: true}
}
