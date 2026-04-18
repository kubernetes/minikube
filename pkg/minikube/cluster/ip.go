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

	"errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/libmachine"
	"k8s.io/minikube/pkg/libmachine/host"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

// HostIP gets the ip address to be used for mapping host -> VM and VM -> host
func HostIP(hostInfo *host.Host, clusterName string) (net.IP, error) {
	switch hostInfo.DriverName {
	case driver.Docker:
		return oci.RoutableHostIPFromInside(oci.Docker, clusterName, hostInfo.Name)
	case driver.Podman:
		return oci.RoutableHostIPFromInside(oci.Podman, clusterName, hostInfo.Name)
	case driver.SSH:
		ip, err := hostInfo.Driver.GetIP()
		if err != nil {
			return []byte{}, fmt.Errorf("Error getting VM/Host IP address: %w", err)
		}
		return net.ParseIP(ip), nil
	case driver.KVM2:
		// `host.Driver.GetIP` returns dhcp lease info for a given network(=`virsh net-dhcp-leases minikube-net`)
		vmIPString, err := hostInfo.Driver.GetIP()
		if err != nil {
			return []byte{}, fmt.Errorf("Error getting VM/Host IP address: %w", err)
		}
		vmIP := net.ParseIP(vmIPString).To4()
		if vmIP == nil {
			// We need the network ip address for minikube-net. It's the start address of the returned subnet.
			return []byte{}, fmt.Errorf("Error converting VM/Host IP address to IPv4 address: %w", err)
		}
		return net.IPv4(vmIP[0], vmIP[1], vmIP[2], byte(1)), nil
	case driver.QEMU, driver.QEMU2:
		ipString, err := hostInfo.Driver.GetIP()
		if err != nil {
			return []byte{}, fmt.Errorf("Error getting IP address: %w", err)
		}
		if ipString == "127.0.0.1" {
			// user network case
			return net.ParseIP("10.0.2.2"), nil
		}
		// socket_vmnet network case
		return net.ParseIP("192.168.105.1"), nil
	case driver.HyperV:
		v := reflect.ValueOf(hostInfo.Driver).Elem()
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
			return []byte{}, fmt.Errorf("%s: %w", fmt.Sprintf("ip for interface (%s)", hypervVirtualSwitch), err)
		}

		return ip, nil
	case driver.VirtualBox:
		vBoxManageCmd := driver.VBoxManagePath()
		out, err := exec.Command(vBoxManageCmd, "showvminfo", hostInfo.Name, "--machinereadable").Output()
		if err != nil {
			return []byte{}, fmt.Errorf("vboxmanage: %w", err)
		}
		// VBox 7.x ARM (darwin/arm64) uses the hostonlynet API, which exposes
		// the attached network as hostonly-network<N>= in showvminfo output
		// (rather than the legacy hostonlyadapter<N>= for hostonlyif).
		// `list hostonlyifs` is empty on that path, so look up `list
		// hostonlynets` and find the matching host-side IP by enumerating
		// local interfaces whose subnet overlaps the hostonlynet.
		// Names written here originate in
		// pkg/drivers/virtualbox.setupHostOnlyNetworkVBox7 which creates
		// hostonlynets with the format "minikube-hostonly-<ip>". This code
		// is tolerant of any name the driver produces; the comment is a
		// reminder for future renames.
		if m := regexp.MustCompile(`hostonly-network2="(.*?)"`).FindStringSubmatch(string(out)); m != nil {
			netName := m[1]
			netList, err := exec.Command(vBoxManageCmd, "list", "hostonlynets").Output()
			if err != nil {
				return []byte{}, fmt.Errorf("Error getting VM/Host IP address: %w", err)
			}
			mask, netAddr, err := parseHostOnlyNet(string(netList), netName)
			if err != nil {
				return []byte{}, fmt.Errorf("hostonlynet %q: %w", netName, err)
			}
			if mask == nil {
				return []byte{}, fmt.Errorf("hostonlynet %q not found in `VBoxManage list hostonlynets` output", netName)
			}
			hostIP, err := findHostIPInSubnet(netAddr, mask)
			if err != nil {
				return []byte{}, fmt.Errorf("host IP for hostonlynet %q: %w", netName, err)
			}
			return hostIP, nil
		}
		re := regexp.MustCompile(`hostonlyadapter2="(.*?)"`)
		iface := re.FindStringSubmatch(string(out))[1]
		ipList, err := exec.Command(vBoxManageCmd, "list", "hostonlyifs").Output()
		if err != nil {
			return []byte{}, fmt.Errorf("Error getting VM/Host IP address: %w", err)
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
			return []byte{}, fmt.Errorf("Error reading the info of Parallels Shared network interface: %w", err)
		}
		re := regexp.MustCompile(`IPv4 address: (.*)`)
		ipMatch := re.FindStringSubmatch(string(out))
		if len(ipMatch) < 2 {
			return []byte{}, fmt.Errorf("Error getting the IP address of Parallels Shared network interface: %w", err)
		}
		ip := ipMatch[1]

		return net.ParseIP(ip), nil
	case driver.HyperKit:
		vmIPString, _ := hostInfo.Driver.GetIP()
		gatewayIPString := vmIPString[:strings.LastIndex(vmIPString, ".")+1] + "1"
		return net.ParseIP(gatewayIPString), nil
	case driver.VMware:
		vmIPString, err := hostInfo.Driver.GetIP()
		if err != nil {
			return []byte{}, fmt.Errorf("Error getting VM IP address: %w", err)
		}
		vmIP := net.ParseIP(vmIPString).To4()
		if vmIP == nil {
			return []byte{}, fmt.Errorf("Error converting VM IP address to IPv4 address: %w", err)
		}
		return net.IPv4(vmIP[0], vmIP[1], vmIP[2], byte(1)), nil
	case driver.VFKit, driver.Krunkit:
		// TODO: check why we need this and test with:
		// - vfkkit+nat
		// - vfkit+vmnet-shared
		// - krunkit+vmnet-shared
		vmIPString, _ := hostInfo.Driver.GetIP()
		gatewayIPString := vmIPString[:strings.LastIndex(vmIPString, ".")+1] + "1"
		return net.ParseIP(gatewayIPString), nil
	case driver.None:
		return net.ParseIP("127.0.0.1"), nil
	default:
		return []byte{}, fmt.Errorf("HostIP not yet implemented for %q driver", hostInfo.DriverName)
	}
}

// DriverIP gets the ip address of the current minikube cluster
func DriverIP(api libmachine.API, machineName string) (net.IP, error) {
	hostInfo, err := machine.LoadHost(api, machineName)
	if err != nil {
		return nil, err
	}

	ipStr, err := hostInfo.Driver.GetIP()
	if err != nil {
		return nil, fmt.Errorf("getting IP: %w", err)
	}
	if driver.IsKIC(hostInfo.DriverName) {
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
		return nil, fmt.Errorf("Could not find interface %s inside %+v", name, ints)
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
	return nil, fmt.Errorf("Unable to find a IPv4 address for interface %q", name)
}

// findHostIPInSubnet returns the first IPv4 address on any local interface
// that lies in the given subnet. Used on darwin/arm64 with the hostonlynet
// API where the host-side IP is auto-assigned by VirtualBox and the actual
// address is only discoverable by enumerating local interfaces.
func findHostIPInSubnet(netAddr net.IP, mask net.IPMask) (net.IP, error) {
	ints, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, in := range ints {
		addrs, err := in.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil {
				continue
			}
			if ip4.Mask(mask).Equal(netAddr) {
				return ip4, nil
			}
		}
	}
	return nil, fmt.Errorf("no local interface found with an IPv4 address in %s/%d", netAddr, maskOnes(mask))
}

func maskOnes(mask net.IPMask) int {
	ones, _ := mask.Size()
	return ones
}

// parseHostOnlyNet extracts the NetworkMask and LowerIP for the named
// hostonlynet from the output of `VBoxManage list hostonlynets`. The
// function is order-tolerant within a record and tolerant of whitespace
// variation. Returns nil, nil if the name isn't found.
//
// Records are delimited by a line starting with "Name:". Blank lines within
// a record (VBoxManage emits one between GUID and State) are ignored.
func parseHostOnlyNet(listOutput, name string) (mask net.IPMask, network net.IP, err error) {
	var blockName, maskStr, lowerStr string
	// finish evaluates the current record against `name` and, if matched,
	// parses and returns the mask/network. It is called when a new record
	// begins (next "Name:" line) and at end-of-input.
	finish := func() (bool, net.IPMask, net.IP, error) {
		if blockName != name {
			return false, nil, nil, nil
		}
		if maskStr == "" || lowerStr == "" {
			return true, nil, nil, fmt.Errorf("hostonlynet %q record missing NetworkMask or LowerIP", name)
		}
		mip := net.ParseIP(maskStr).To4()
		if mip == nil {
			return true, nil, nil, fmt.Errorf("hostonlynet %q: unable to parse NetworkMask %q", name, maskStr)
		}
		lip := net.ParseIP(lowerStr).To4()
		if lip == nil {
			return true, nil, nil, fmt.Errorf("hostonlynet %q: unable to parse LowerIP %q", name, lowerStr)
		}
		m := net.IPMask(mip)
		return true, m, lip.Mask(m), nil
	}

	for _, line := range strings.Split(listOutput, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "Name":
			// Starting a new record: evaluate the previous one.
			if blockName != "" {
				if matched, m, n, err := finish(); matched {
					return m, n, err
				}
			}
			blockName, maskStr, lowerStr = val, "", ""
		case "NetworkMask":
			maskStr = val
		case "LowerIP":
			lowerStr = val
		}
	}
	// Evaluate the last record.
	if blockName != "" {
		if matched, m, n, err := finish(); matched {
			return m, n, err
		}
	}
	return nil, nil, nil
}
