/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package network

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

var (
	// valid private network subnets (RFC1918)
	privateSubnets = []net.IPNet{
		// 10.0.0.0/8
		{
			IP:   []byte{10, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		// 172.16.0.0/12
		{
			IP:   []byte{172, 16, 0, 0},
			Mask: []byte{255, 240, 0, 0},
		},
		// 192.168.0.0/16
		{
			IP:   []byte{192, 168, 0, 0},
			Mask: []byte{255, 255, 0, 0},
		},
	}
)

// Parameters contains main network parameters.
type Parameters struct {
	IP        string // IP address of the network
	Netmask   string // form: 4-byte ('a.b.c.d')
	CIDR      string // form: CIDR
	Gateway   string // first IP address (assumed, not checked !)
	ClientMin string // second IP address
	ClientMax string // last IP address before broadcastS
	Broadcast string // last IP address
	Interface
}

// Interface contains main network interface parameters.
type Interface struct {
	IfaceName string
	IfaceIPv4 string
	IfaceMTU  int
	IfaceMAC  string
}

// Inspect initialises IPv4 network parameters struct from given address.
// address can be single address (like "192.168.17.42"), network address (like "192.168.17.0"), or in cidr form (like "192.168.17.42/24 or "192.168.17.0/24").
// If addr is valid existsing interface address, network struct will also contain info about the respective interface.
func Inspect(addr string) (*Parameters, error) {
	n := &Parameters{}

	// extract ip from addr
	ip, network, err := net.ParseCIDR(addr)
	if err != nil {
		ip = net.ParseIP(addr)
		if ip == nil {
			return nil, errors.Wrapf(err, "parsing address %q", addr)
		}
	}

	// check local interfaces
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		ifAddrs, err := iface.Addrs()
		if err != nil {
			return nil, errors.Wrapf(err, "listing addresses of network interface %+v", iface)
		}
		for _, ifAddr := range ifAddrs {
			ifip, lan, err := net.ParseCIDR(ifAddr.String())
			if err != nil {
				return nil, errors.Wrapf(err, "parsing address of network iface %+v", ifAddr)
			}
			if lan.Contains(ip) {
				n.IfaceName = iface.Name
				n.IfaceIPv4 = ifip.To4().String()
				n.IfaceMTU = iface.MTU
				n.IfaceMAC = iface.HardwareAddr.String()
				n.Gateway = n.IfaceIPv4
				network = lan
				break
			}
		}
	}

	if network == nil {
		ipnet := &net.IPNet{
			IP:   ip,
			Mask: ip.DefaultMask(), // assume default network mask
		}
		_, network, err = net.ParseCIDR(ipnet.String())
		if err != nil {
			return nil, errors.Wrapf(err, "determining network address from %q", addr)
		}
	}

	n.IP = network.IP.String()
	n.Netmask = net.IP(network.Mask).String() // form: 4-byte ('a.b.c.d')
	n.CIDR = network.String()

	networkIP := binary.BigEndian.Uint32(network.IP)                      // IP address of the network
	networkMask := binary.BigEndian.Uint32(network.Mask)                  // network mask
	broadcastIP := (networkIP & networkMask) | (networkMask ^ 0xffffffff) // last network IP address

	broadcast := make(net.IP, 4)
	binary.BigEndian.PutUint32(broadcast, broadcastIP)
	n.Broadcast = broadcast.String()

	gateway := net.ParseIP(n.Gateway).To4() // has to be converted to 4-byte representation!
	if gateway == nil {
		gateway = make(net.IP, 4)
		binary.BigEndian.PutUint32(gateway, networkIP+1) // assume first network IP address
		n.Gateway = gateway.String()
	}
	gatewayIP := binary.BigEndian.Uint32(gateway)

	min := make(net.IP, 4)
	binary.BigEndian.PutUint32(min, gatewayIP+1) // clients-from: first network IP address after gateway
	n.ClientMin = min.String()

	max := make(net.IP, 4)
	binary.BigEndian.PutUint32(max, broadcastIP-1) // clients-from: last network IP address before broadcast
	n.ClientMax = max.String()

	return n, nil
}

// IsSubnetTaken returns if local network subnet exists and any error occurred.
// If will return false in case of an error.
func IsSubnetTaken(subnet string) (bool, error) {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		return false, errors.Wrap(err, "listing local networks")
	}
	for _, ip := range ips {
		_, lan, err := net.ParseCIDR(ip.String())
		if err != nil {
			return false, errors.Wrapf(err, "parsing network iface address %q", ip)
		}
		if lan.Contains(net.ParseIP(subnet)) {
			return true, nil
		}
	}
	return false, nil
}

// IsSubnetPrivate returns if subnet is a private network.
func IsSubnetPrivate(subnet string) bool {
	for _, ipnet := range privateSubnets {
		if ipnet.Contains(net.ParseIP(subnet)) {
			return true
		}
	}
	return false
}

// FreeSubnet will try to find free private network beginning with startSubnet, incrementing it in steps up to number of tries.
func FreeSubnet(startSubnet string, step, tries int) (*Parameters, error) {
	for try := 0; try < tries; try++ {
		n, err := Inspect(startSubnet)
		if err != nil {
			return nil, err
		}
		startSubnet = n.IP
		if IsSubnetPrivate(startSubnet) {
			taken, err := IsSubnetTaken(startSubnet)
			if err != nil {
				return nil, err
			}
			if !taken {
				klog.Infof("using free private subnet %s: %+v", n.CIDR, n)
				return n, nil
			}
			klog.Infof("skipping subnet %s that is taken: %+v", n.CIDR, n)
		} else {
			klog.Infof("skipping subnet %s that is not private", n.CIDR)
		}
		ones, _ := net.ParseIP(n.IP).DefaultMask().Size()
		nextSubnet := net.ParseIP(startSubnet).To4()
		if ones <= 16 {
			nextSubnet[1] += byte(step)
		} else {
			nextSubnet[2] += byte(step)
		}
		startSubnet = nextSubnet.String()
	}
	return nil, fmt.Errorf("no free private network subnets found with given parameters")
}
