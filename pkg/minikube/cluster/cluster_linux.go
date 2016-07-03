/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"fmt"
	"path/filepath"

	"github.com/dhiltgen/docker-machine-kvm"
	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/minikube/constants"
)

func createKVMHost(config MachineConfig) drivers.Driver {
	d := kvm.NewDriver(constants.MachineName, constants.Minipath).(*kvm.Driver)
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.Network = "default"
	d.PrivateNetwork = "docker-machines"
	d.Boot2DockerURL = config.MinikubeISO
	d.DiskSize = 20000
	d.DiskPath = filepath.Join(constants.Minipath, "machines", constants.MachineName, fmt.Sprintf("%s.img", constants.MachineName))
	d.ISO = filepath.Join(constants.Minipath, "machines", constants.MachineName, "boot2docker.iso")
	d.CacheMode = "default"
	d.IOMode = "threads"
	return d
}
