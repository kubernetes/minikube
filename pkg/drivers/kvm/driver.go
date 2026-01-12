//go:build linux

/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package kvm

import (
	"k8s.io/minikube/pkg/drivers/common"
	"k8s.io/minikube/pkg/libmachine/drivers"
)

const (
	qemusystem                = "qemu:///system"
	defaultPrivateNetworkName = "minikube-net"
	defaultNetworkName        = "default"
)

// Driver is the machine driver for KVM
type Driver struct {
	*drivers.BaseDriver
	*common.CommonDriver

	// How much memory, in MB, to allocate to the VM
	Memory int

	// How many cpus to allocate to the VM
	CPU int

	// The name of the default network
	Network string

	// The name of the private network
	PrivateNetwork string

	// The size of the disk to be created for the VM, in MB
	DiskSize int

	// The path of the disk .img
	DiskPath string

	// A file or network URI to fetch the minikube ISO
	Boot2DockerURL string

	// The location of the iso to boot from
	ISO string

	// The randomly generated MAC Address
	// If empty, a random MAC will be generated.
	MAC string

	// The randomly generated MAC Address for the NIC attached to the private network
	// If empty, a random MAC will be generated.
	PrivateMAC string

	// Whether to passthrough GPU devices from the host to the VM.
	GPU bool

	// Whether to hide the KVM hypervisor signature from the guest
	Hidden bool

	// XML that needs to be added to passthrough GPU devices.
	DevicesXML string

	// QEMU Connection URI
	ConnectionURI string

	// NUMA node count default value is 1
	NUMANodeCount int

	// NUMA XML
	NUMANodeXML string

	// Extra Disks
	ExtraDisks int

	// Extra Disks XML
	ExtraDisksXML []string
}

// NewDriver creates a new driver for a host
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
			SSHUser:     "docker",
		},
		CommonDriver:   &common.CommonDriver{},
		PrivateNetwork: defaultPrivateNetworkName,
		Network:        defaultNetworkName,
		ConnectionURI:  qemusystem,
	}
}
