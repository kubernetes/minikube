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
	"errors"
	"fmt"
	"net"
	"path/filepath"

	kvm "github.com/dhiltgen/docker-machine-kvm"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"k8s.io/minikube/pkg/minikube/constants"
)

func createKVMHost(config MachineConfig) *kvm.Driver {
	return &kvm.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: constants.MachineName,
			StorePath:   constants.GetMinipath(),
		},
		Memory:         config.Memory,
		CPU:            config.CPUs,
		Network:        config.KvmNetwork,
		PrivateNetwork: "docker-machines",
		Boot2DockerURL: config.Downloader.GetISOFileURI(config.MinikubeISO),
		DiskSize:       config.DiskSize,
		DiskPath:       filepath.Join(constants.GetMinipath(), "machines", constants.MachineName, fmt.Sprintf("%s.img", constants.MachineName)),
		ISO:            filepath.Join(constants.GetMinipath(), "machines", constants.MachineName, "boot2docker.iso"),
		CacheMode:      "default",
		IOMode:         "threads",
	}
}

func getVMHostIP(host *host.Host) (net.IP, error) {
	switch host.DriverName {
	case "virtualbox":
		return net.ParseIP("10.0.2.2"), nil
	case "kvm":
		return net.ParseIP("10.0.2.2"), nil
	default:
		return []byte{}, errors.New("Error, attempted to get host ip address for unsupported driver")
	}
}
