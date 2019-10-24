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
	"fmt"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

const defaultVirtualboxNicType = "virtio"

func init() {
	err := registry.Register(registry.DriverDef{
		Name:          driver.VirtualBox,
		Builtin:       true,
		ConfigCreator: createVirtualboxHost,
		DriverCreator: func() drivers.Driver {
			return virtualbox.NewDriver("", "")
		},
	})
	if err != nil {
		panic(fmt.Sprintf("unable to register: %v", err))
	}
}

func createVirtualboxHost(mc config.MachineConfig) interface{} {
	d := virtualbox.NewDriver(config.GetMachineName(), localpath.MiniPath())

	d.Boot2DockerURL = mc.Downloader.GetISOFileURI(mc.MinikubeISO)
	d.Memory = mc.Memory
	d.CPU = mc.CPUs
	d.DiskSize = mc.DiskSize
	d.HostOnlyCIDR = mc.HostOnlyCIDR
	d.NoShare = mc.DisableDriverMounts
	d.NoVTXCheck = mc.NoVTXCheck
	d.NatNicType = defaultVirtualboxNicType
	d.HostOnlyNicType = defaultVirtualboxNicType
	d.DNSProxy = mc.DNSProxy
	d.HostDNSResolver = mc.HostDNSResolver

	return d
}
