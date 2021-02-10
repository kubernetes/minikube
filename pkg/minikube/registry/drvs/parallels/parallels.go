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
	"os/exec"

	parallels "github.com/Parallels/docker-machine-parallels/v2"
	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	err := registry.Register(registry.DriverDef{
		Name:     driver.Parallels,
		Config:   configure,
		Status:   status,
		Priority: registry.Default,
		Init:     func() drivers.Driver { return parallels.NewDriver("", "") },
	})
	if err != nil {
		panic(fmt.Sprintf("unable to register: %v", err))
	}

}

func configure(cfg config.ClusterConfig, n config.Node) (interface{}, error) {
	d := parallels.NewDriver(config.MachineName(cfg, n), localpath.MiniPath()).(*parallels.Driver)
	d.Boot2DockerURL = download.LocalISOResource(cfg.MinikubeISO)
	d.Memory = cfg.Memory
	d.CPU = cfg.CPUs
	d.DiskSize = cfg.DiskSize
	return d, nil
}

func status() registry.State {
	_, err := exec.LookPath("prlctl")
	if err != nil {
		return registry.State{Error: err, Fix: "Install Parallels Desktop for Mac", Doc: "https://minikube.sigs.k8s.io/docs/drivers/parallels/"}
	}
	return registry.State{Installed: true, Healthy: true}
}
