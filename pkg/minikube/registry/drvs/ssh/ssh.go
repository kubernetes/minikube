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

package ssh

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/drivers/ssh"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	err := registry.Register(registry.DriverDef{
		Name:     driver.SSH,
		Alias:    []string{driver.AliasSSH},
		Config:   configure,
		Status:   status,
		Priority: registry.Discouraged, // requires external VM
		Init:     func() drivers.Driver { return ssh.NewDriver(ssh.Config{}) },
	})
	if err != nil {
		panic(fmt.Sprintf("unable to register: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	d := ssh.NewDriver(ssh.Config{
		MachineName:      config.MachineName(cc, n),
		StorePath:        localpath.MiniPath(),
		ContainerRuntime: cc.KubernetesConfig.ContainerRuntime,
	})

	if cc.SSHIPAddress == "" {
		return nil, errors.Errorf("please provide an IP address")
	}

	// We don't want the API server listening on loopback interface,
	// even if we might use a tunneled VM port for the SSH service
	if cc.SSHIPAddress == "127.0.0.1" || cc.SSHIPAddress == "localhost" {
		return nil, errors.Errorf("please provide real IP address")
	}

	d.IPAddress = cc.SSHIPAddress
	d.SSHUser = cc.SSHUser
	d.SSHKey = cc.SSHKey
	d.SSHPort = cc.SSHPort

	return d, nil
}

func status() registry.State {
	return registry.State{Installed: true, Healthy: true}
}
