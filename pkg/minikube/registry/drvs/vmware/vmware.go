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

package vmware

import (
	"fmt"
	"os/exec"

	vmwcfg "github.com/machine-drivers/docker-machine-driver-vmware/pkg/drivers/vmware/config"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	err := registry.Register(registry.DriverDef{
		Name:     driver.VMware,
		Config:   configure,
		Priority: registry.Default,
		Status:   status,
	})
	if err != nil {
		panic(fmt.Sprintf("unable to register: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	d := vmwcfg.NewConfig(config.MachineName(cc, n), localpath.MiniPath())
	d.Boot2DockerURL = download.LocalISOResource(cc.MinikubeISO)
	d.Memory = cc.Memory
	d.CPU = cc.CPUs
	d.DiskSize = cc.DiskSize

	// TODO(frapposelli): push these defaults upstream to fixup this driver
	d.SSHPort = 22
	d.ISO = d.ResolveStorePath("boot2docker.iso")
	return d, nil
}

func status() registry.State {
	_, err := exec.LookPath("docker-machine-driver-vmware")
	if err != nil {
		return registry.State{Error: err, Fix: "Install docker-machine-driver-vmware", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/vmware/"}
	}
	_, err = exec.LookPath("vmrun")
	if err != nil {
		return registry.State{Error: err, Fix: "Install vmrun", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/vmware/"}
	}
	return registry.State{Installed: true, Healthy: true}
}
