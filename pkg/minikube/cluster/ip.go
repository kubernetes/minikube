/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"os/exec"
	"regexp"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

// GetVMHostIP gets the ip address to be used for mapping host -> VM and VM -> host
func GetVMHostIP(host *host.Host) (net.IP, error) {
	switch host.DriverName {
	case driver.Docker:
		return oci.RoutableHostIPFromInside(oci.Docker, host.Name)
	case driver.Podman:
		return oci.RoutableHostIPFromInside(oci.Podman, host.Name)
	case driver.KVM2:
		return net.ParseIP("192.168.39.1"), nil
	case driver.HyperV:
		re := regexp.MustCompile(`"VSwitch": "(.*?)",`)
		// TODO(aprindle) Change this to deserialize the driver instead
		hypervVirtualSwitch := re.FindStringSubmatch(string(host.RawDriver))[1]
		ip, err := getIPForInterface(fmt.Sprintf("vEthernet (%s)", hypervVirtualSwitch))
		if err != nil {
			return []byte{}, errors.Wrap(err, fmt.Sprintf("ip for interface (%s)", hypervVirtualSwitch))
		}
		return ip, nil
	case driver.VirtualBox:
		out, err := exec.Command(driver.VBoxManagePath(), "showvminfo", host.Name, "--machinereadable").Output()
		if err != nil {
			return []byte{}, errors.Wrap(err, "vboxmanage")
		}
		re := regexp.MustCompile(`hostonlyadapter2="(.*?)"`)
		iface := re.FindStringSubmatch(string(out))[1]
		ip, err := getIPForInterface(iface)
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting VM/Host IP address")
		}
		return ip, nil
	case driver.HyperKit:
		return net.ParseIP("192.168.64.1"), nil
	case driver.VMware:
		vmIPString, err := host.Driver.GetIP()
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting VM IP address")
		}
		vmIP := net.ParseIP(vmIPString).To4()
		if vmIP == nil {
			return []byte{}, errors.Wrap(err, "Error converting VM IP address to IPv4 address")
		}
		return net.IPv4(vmIP[0], vmIP[1], vmIP[2], byte(1)), nil
	default:
		return []byte{}, errors.New("Error, attempted to get host ip address for unsupported driver")
	}
}

// GetHostDriverIP gets the ip address of the current minikube cluster
func GetHostDriverIP(api libmachine.API, machineName string) (net.IP, error) {
	host, err := machine.CheckIfHostExistsAndLoad(api, machineName)
	if err != nil {
		return nil, err
	}

	ipStr, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrap(err, "getting IP")
	}
	if driver.IsKIC(host.DriverName) {
		ipStr = oci.DefaultBindIPV4
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("parsing IP: %s", ipStr)
	}
	return ip, nil
}

// Based on code from http://stackoverflow.com/questions/23529663/how-to-get-all-addresses-and-masks-from-local-interfaces-in-go
func getIPForInterface(name string) (net.IP, error) {
	i, _ := net.InterfaceByName(name)
	addrs, _ := i.Addrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip := ipnet.IP.To4(); ip != nil {
				return ip, nil
			}
		}
	}
	return nil, errors.Errorf("Error finding IPV4 address for %s", name)
}
