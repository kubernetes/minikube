/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"net"
)

// hostOnlyNet represents a VirtualBox 7.x "Host Only Network" (distinct from
// the legacy "host-only interface" managed via `hostonlyif`). The hostonlynet
// API replaces the kernel-extension-based hostonlyif on modern macOS and is
// required on darwin/arm64.
type hostOnlyNet struct {
	Name            string
	GUID            string
	Enabled         bool
	NetworkMask     net.IPMask
	LowerIP         net.IP
	UpperIP         net.IP
	VBoxNetworkName string
}

// listHostOnlyNets parses the output of `VBoxManage list hostonlynets`.
func listHostOnlyNets(vbox VBoxManager) (map[string]*hostOnlyNet, error) {
	out, err := vbox.vbmOut("list", "hostonlynets")
	if err != nil {
		return nil, err
	}

	byName := map[string]*hostOnlyNet{}
	n := &hostOnlyNet{}

	err = parseKeyValues(out, reColonLine, func(key, val string) error {
		switch key {
		case "Name":
			n.Name = val
		case "GUID":
			n.GUID = val
		case "State":
			n.Enabled = val == "Enabled"
		case "NetworkMask":
			n.NetworkMask = parseIPv4Mask(val)
		case "LowerIP":
			n.LowerIP = net.ParseIP(val)
		case "UpperIP":
			n.UpperIP = net.ParseIP(val)
		case "VBoxNetworkName":
			n.VBoxNetworkName = val
			if _, present := byName[n.Name]; present {
				return fmt.Errorf("VirtualBox has multiple host-only networks with the same name %q", n.Name)
			}
			byName[n.Name] = n
			n = &hostOnlyNet{}
		default:
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return byName, nil
}

// findHostOnlyNetByRange searches for an existing hostonlynet whose IP range
// covers the given host IP in the given network. Returns nil if not found.
func findHostOnlyNetByRange(nets map[string]*hostOnlyNet, hostIP net.IP, netmask net.IPMask) *hostOnlyNet {
	for _, n := range nets {
		if n.NetworkMask.String() == netmask.String() &&
			n.LowerIP != nil && n.UpperIP != nil &&
			ipInRange(hostIP, n.LowerIP, n.UpperIP) {
			return n
		}
	}
	return nil
}

// ipInRange reports whether ip is between lower and upper inclusive.
// Assumes all three are IPv4.
func ipInRange(ip, lower, upper net.IP) bool {
	ip4 := ip.To4()
	lo4 := lower.To4()
	up4 := upper.To4()
	if ip4 == nil || lo4 == nil || up4 == nil {
		return false
	}
	return bytesCompare4(ip4, lo4) >= 0 && bytesCompare4(ip4, up4) <= 0
}

func bytesCompare4(a, b net.IP) int {
	for i := 0; i < 4; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

// createHostOnlyNet creates a new hostonlynet with the given name, netmask,
// and IP range. Returns the created hostonlynet populated from a follow-up
// list call.
func createHostOnlyNet(vbox VBoxManager, name string, netmask net.IPMask, lowerIP, upperIP net.IP) (*hostOnlyNet, error) {
	if err := vbox.vbm(
		"hostonlynet", "add",
		"--name", name,
		"--netmask", net.IP(netmask).String(),
		"--lower-ip", lowerIP.String(),
		"--upper-ip", upperIP.String(),
		"--enable",
	); err != nil {
		return nil, fmt.Errorf("hostonlynet add: %w", err)
	}

	nets, err := listHostOnlyNets(vbox)
	if err != nil {
		return nil, err
	}
	n, ok := nets[name]
	if !ok {
		return nil, fmt.Errorf("hostonlynet %q created but not found in list", name)
	}
	return n, nil
}
