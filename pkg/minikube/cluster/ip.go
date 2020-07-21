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
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

// HostIP gets the ip address to be used for mapping host -> VM and VM -> host
func HostIP(host *host.Host) (net.IP, error) {
	switch host.DriverName {
	case driver.Docker:
		return oci.RoutableHostIPFromInside(oci.Docker, host.Name)
	case driver.Podman:
		return oci.RoutableHostIPFromInside(oci.Podman, host.Name)
	case driver.KVM2:
		return net.ParseIP("192.168.39.1"), nil
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
	ints, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var i net.Interface
	for _, in := range ints {
		if strings.HasPrefix(in.Name, name) {
			i = in
			break
		}
	}

	// Didn't find prefix, let's try any substring
	if i.Name == "" {
		for _, in := range ints {
			if strings.Contains(in.Name, name) {
				i = in
				break
			}
		}
	}

	// We found nothing, fail out
	if i.Name == "" {
		return nil, errors.Errorf("Could not find interface %s inside %+v", name, ints)
	}

	glog.Infof("Found hyperv interface: %+v\n", i)
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
