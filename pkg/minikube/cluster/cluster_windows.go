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
	"net"
	"regexp"
	"strings"

	"github.com/docker/machine/drivers/hyperv"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
)

func createHypervHost(config MachineConfig) drivers.Driver {
	d := hyperv.NewDriver(constants.MachineName, constants.GetMinipath())
	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.VSwitch = config.HypervVirtualSwitch
	d.MemSize = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = int(config.DiskSize)
	d.SSHUser = "docker"
	return d
}

func getVMHostIP(host *host.Host) (net.IP, error) {
	switch host.DriverName {
	case "virtualbox":
		return net.ParseIP("10.0.2.2"), nil
	case "hyperv":
		re := regexp.MustCompile("\"VSwitch\": \"(.*?)\",")
		hypervVirtualSwitch := re.FindStringSubmatch(string(host.RawDriver))[1]
		ip, err := getWindowsHostIpFromHyperV(hypervVirtualSwitch)
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting 9p mount command")
		}
		return ip, nil
	default:
		return []byte{}, errors.New("Error, attempted to get host ip address for unsupported driver")
	}
}

func getWindowsHostIpFromHyperV(hypervVirtualSwitch string) (net.IP, error) {
	virtualSwitchTemplate := "vEthernet (%s)"

	i, _ := net.InterfaceByName(fmt.Sprintf(virtualSwitchTemplate, hypervVirtualSwitch))
	addrs, _ := i.Addrs()
	for _, a := range addrs {
		switch a.(type) {
		case *net.IPNet:
			ip := a.String()
			if strings.Contains(ip, ".") {
				vmIP := net.ParseIP(strings.Split(ip, "/")[0])
				if vmIP.String() == "" {
					return nil, errors.Errorf("Error finding IPV4 address for virtual switch %s", hypervVirtualSwitch)
				}
				return vmIP, nil
			}
		}
	}
	return nil, errors.Errorf("Error finding IPV4 address for virtual switch %s", hypervVirtualSwitch)
}
