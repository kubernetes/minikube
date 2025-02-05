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
	"reflect"
	"regexp"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

// HostIP gets the ip address to be used for mapping host -> VM and VM -> host
func HostIP(host *host.Host, clusterName string) (net.IP, error) {
	switch host.DriverName {
	case driver.Docker:
		return oci.RoutableHostIPFromInside(oci.Docker, clusterName, host.Name)
	case driver.Podman:
		return oci.RoutableHostIPFromInside(oci.Podman, clusterName, host.Name)
	case driver.SSH:
		ip, err := host.Driver.GetIP()
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting VM/Host IP address")
		}
		return net.ParseIP(ip), nil
	case driver.KVM2:
		// `host.Driver.GetIP` returns dhcp lease info for a given network(=`virsh net-dhcp-leases minikube-net`)
		vmIPString, err := host.Driver.GetIP()
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting VM/Host IP address")
		}
		vmIP := net.ParseIP(vmIPString).To4()
		if vmIP == nil {
			// We need the network ip address for minikube-net. It's the start address of the returned subnet.
			return []byte{}, errors.Wrap(err, "Error converting VM/Host IP address to IPv4 address")
		}
		return net.IPv4(vmIP[0], vmIP[1], vmIP[2], byte(1)), nil
	case driver.QEMU, driver.QEMU2:
		ipString, err := host.Driver.GetIP()
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting IP address")
		}
		if ipString == "127.0.0.1" {
			// user network case
			return net.ParseIP("10.0.2.2"), nil
		}
		// "socket_vmnet" and "vmnet-*" network case: host ip should be start address of subnet
		vmIP := net.ParseIP(ipString).To4()
		if vmIP == nil {
			return []byte{}, errors.Wrap(err, "Error converting VM IP address to IPv4 address")
		}
		return net.IPv4(vmIP[0], vmIP[1], vmIP[2], byte(1)), nil
	case driver.HyperV:
		v := reflect.ValueOf(host.Driver).Elem()
		var hypervVirtualSwitch string
		// We don't have direct access to hyperv.Driver so use reflection to retrieve the virtual switch name
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Name == "VSwitch" {
				hypervVirtualSwitch = v.Field(i).Interface().(string)

				break
			}
		}
		if hypervVirtualSwitch == "" {
			return nil, errors.New("No virtual switch found")
		}
		ip, err := getIPForInterface(fmt.Sprintf("vEthernet (%s)", hypervVirtualSwitch))
		if err != nil {
			return []byte{}, errors.Wrap(err, fmt.Sprintf("ip for interface (%s)", hypervVirtualSwitch))
		}

		return ip, nil
	case driver.VirtualBox:
		vBoxManageCmd := driver.VBoxManagePath()
		out, err := exec.Command(vBoxManageCmd, "showvminfo", host.Name, "--machinereadable").Output()
		if err != nil {
			return []byte{}, errors.Wrap(err, "vboxmanage")
		}
		re := regexp.MustCompile(`hostonlyadapter2="(.*?)"`)
		iface := re.FindStringSubmatch(string(out))[1]
		ipList, err := exec.Command(vBoxManageCmd, "list", "hostonlyifs").Output()
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting VM/Host IP address")
		}
		re = regexp.MustCompile(`(?sm)Name:\s*` + iface + `\s*$.+?IPAddress:\s*(\S+)`)
		ip := re.FindStringSubmatch(string(ipList))[1]

		return net.ParseIP(ip), nil
	case driver.Parallels:
		bin := "prlsrvctl"
		var binPath string
		if fullPath, err := exec.LookPath(bin); err != nil {
			binPath = fullPath
		} else {
			binPath = bin
		}
		out, err := exec.Command(binPath, "net", "info", "Shared").Output()
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error reading the info of Parallels Shared network interface")
		}
		re := regexp.MustCompile(`IPv4 address: (.*)`)
		ipMatch := re.FindStringSubmatch(string(out))
		if len(ipMatch) < 2 {
			return []byte{}, errors.Wrap(err, "Error getting the IP address of Parallels Shared network interface")
		}
		ip := ipMatch[1]

		return net.ParseIP(ip), nil
	case driver.HyperKit:
		vmIPString, _ := host.Driver.GetIP()
		gatewayIPString := vmIPString[:strings.LastIndex(vmIPString, ".")+1] + "1"
		return net.ParseIP(gatewayIPString), nil
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
	case driver.VFKit:
		vmIPString, _ := host.Driver.GetIP()
		gatewayIPString := vmIPString[:strings.LastIndex(vmIPString, ".")+1] + "1"
		return net.ParseIP(gatewayIPString), nil
	case driver.None:
		return net.ParseIP("127.0.0.1"), nil
	default:
		return []byte{}, fmt.Errorf("HostIP not yet implemented for %q driver", host.DriverName)
	}
}

// DriverIP gets the ip address of the current minikube cluster
func DriverIP(api libmachine.API, machineName string) (net.IP, error) {
	host, err := machine.LoadHost(api, machineName)
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
	klog.Infof("getIPForInterface: searching for %q", name)
	ints, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var i net.Interface
	for _, in := range ints {
		if strings.HasPrefix(strings.ToLower(in.Name), strings.ToLower(name)) {
			klog.Infof("found prefix matching interface for %q: %q", name, in.Name)
			i = in

			break
		}
		klog.Infof("%q does not match prefix %q", in.Name, name)
	}

	// Didn't find prefix, let's try any substring
	if i.Name == "" {
		for _, in := range ints {
			if strings.Contains(strings.ToLower(in.Name), strings.ToLower(name)) {
				klog.Infof("found substring matching interface for %q: %q", name, in.Name)
				i = in

				break
			}
			klog.Infof("%q does not match substring %q", in.Name, name)
		}
	}

	// We found nothing, fail out
	if i.Name == "" {
		return nil, errors.Errorf("Could not find interface %s inside %+v", name, ints)
	}

	klog.Infof("Found interface: %+v\n", i)
	addrs, _ := i.Addrs()
	for _, a := range addrs {
		klog.Infof("interface addr: %+v", a)
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip := ipnet.IP.To4(); ip != nil {
				return ip, nil
			}
		}
	}
	return nil, errors.Errorf("Unable to find a IPv4 address for interface %q", name)
}
