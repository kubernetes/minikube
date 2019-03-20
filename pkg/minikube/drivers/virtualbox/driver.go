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

package virtualbox

import (
	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine/drivers"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

const defaultVirtualboxNicType = "virtio"

func init() {
	registry.Register(registry.DriverDef{
		Name:          "virtualbox",
		Builtin:       true,
		ConfigCreator: createVirtualboxHost,
		DriverCreator: func() drivers.Driver {
			return virtualbox.NewDriver("", "")
		},
	})
}

func createVirtualboxHost(config cfg.MachineConfig) interface{} {
	d := virtualbox.NewDriver(cfg.GetMachineName(), constants.GetMinipath())

	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = int(config.DiskSize)
	d.HostOnlyCIDR = config.HostOnlyCIDR
	d.NoShare = config.DisableDriverMounts
	d.NoVTXCheck = config.NoVTXCheck
	d.NatNicType = defaultVirtualboxNicType
	d.HostOnlyNicType = defaultVirtualboxNicType
	d.DNSProxy = false
	d.HostDNSResolver = true

	return d
}
