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

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnutils"
	"k8s.io/minikube/pkg/minikube/cluster/local"
	"k8s.io/minikube/pkg/minikube/constants"
)

type kvmDriver struct {
	*drivers.BaseDriver

	Memory         int
	DiskSize       int
	CPU            int
	Network        string
	PrivateNetwork string
	ISO            string
	Boot2DockerURL string
	DiskPath       string
	CacheMode      string
	IOMode         string
}

func createKVMHost(config MachineConfig) *kvmDriver {
	return &kvmDriver{
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
		return net.ParseIP("192.168.42.1"), nil
	default:
		return []byte{}, errors.New("Error, attempted to get host ip address for unsupported driver")
	}
}

func createLocalHost(config MachineConfig) *local.Driver {
	return &local.Driver{
		EnginePort: engine.DefaultPort,
		BaseDriver: &drivers.BaseDriver{
			MachineName: constants.MachineName,
			StorePath:   constants.GetMinipath(),
			IPAddress:   "127.0.0.1",
			SSHUser:     mcnutils.GetUsername(),
			SSHPort:     22,
			SSHKeyPath:  filepath.Join(mcnutils.GetHomeDir(), ".ssh", "id_rsa"),
		},
	}
}
