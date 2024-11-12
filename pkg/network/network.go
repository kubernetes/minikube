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
	"time"

	"github.com/juju/mutex/v2"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/util/lock"
)

// Parameters contains main network parameters.
type Parameters struct {
	IP        string // IP address of network
	Netmask   string // dotted-decimal format ('a.b.c.d')
	Prefix    int    // network prefix length (number of leading ones in network mask)
	CIDR      string // CIDR format ('a.b.c.d/n')
	Gateway   string // taken from network interface address or assumed as first network IP address from given addr
	ClientMin string // first available client IP address after gateway
	ClientMax string // last available client IP address before broadcast
	Broadcast string // last network IP address
	IsPrivate bool   // whether the IP is private or not
	Interface
	reservation mutex.Releaser // subnet reservation has lifespan of the process: "If a process dies while the mutex is held, the mutex is automatically released."
}

// Interface contains main network interface parameters.
type Interface struct {
	IfaceName string
	IfaceIPv4 string
	IfaceMTU  int
	IfaceMAC  string
}

// lookupInInterfaces iterates over all local network interfaces
// and tries to match "ip" with associated networks
// returns (network parameters, ip network, nil) if found
//
//	(nil, nil, nil) it nof
//	(nil, nil, error) if any error happened
func lookupInInterfaces(ip net.IP) (*Parameters, *net.IPNet, error) {
	// check local network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, fmt.Errorf("failed listing network interfaces: %w", err)
	}

	for _, iface := range ifaces {

		ifAddrs, err := iface.Addrs()
		if err != nil {
			return nil, nil, fmt.Errorf("failed listing addresses of network interface %+v: %w", iface, err)
		}

		for _, ifAddr := range ifAddrs {
			ifip, lan, err := net.ParseCIDR(ifAddr.String())
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing network interface address %+v: %w", ifAddr, err)
			}
			if lan.Contains(ip) {
				ip4 := ifip.To4().String()
				rt := Parameters{
					Interface: Interface{
						IfaceName: iface.Name,
						IfaceIPv4: ip4,
						IfaceMTU:  iface.MTU,
						IfaceMAC:  iface.HardwareAddr.String(),
					},
					Gateway: ip4,
				}
				return &rt, lan, nil
			}
		}
	}
	return nil, nil, nil
}

// Inspect initialises IPv4 network parameters struct from given address addr.
// addr can be single address (like "192.168.17.42"), network address (like "192.168.17.0") or in CIDR form (like "192.168.17.42/24 or "192.168.17.0/24").
// If addr belongs to network of local network interface, parameters will also contain info about that network interface.
var Inspect = func(addr string) (*Parameters, error) {
	// extract ip from addr
	ip, network, err := ParseAddr(addr)
	if err != nil {
		return nil, err
	}

	n := &Parameters{}

	ifParams, ifNet, err := lookupInInterfaces(ip)
	if err != nil {
		return nil, err
	}
	if ifNet != nil {
		network = ifNet
		n = ifParams
	}

	// couldn't determine network parameters from addr nor from network interfaces
	if network == nil {
		ipnet := &net.IPNet{
			IP:   ip,
			Mask: ip.DefaultMask(), // assume default network mask
		}
		_, network, err = net.ParseCIDR(ipnet.String())
		if err != nil {
			return nil, fmt.Errorf("failed determining address of network from %s: %w", addr, err)
		}
	}

	n.IP = network.IP.String()
	n.Netmask = net.IP(network.Mask).String() // dotted-decimal format ('a.b.c.d')
	n.Prefix, _ = network.Mask.Size()
	n.CIDR = network.String()
	n.IsPrivate = network.IP.IsPrivate()

	networkIP := binary.BigEndian.Uint32(network.IP)                      // IP address of network
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

	minIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(minIP, gatewayIP+1) // clients-from: first network IP address after gateway
	n.ClientMin = minIP.String()

	maxIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(maxIP, broadcastIP-1) // clients-to: last network IP address before broadcast
	n.ClientMax = maxIP.String()

	return n, nil
}

// isSubnetTaken returns if local network subnet exists and any error occurred.
// If will return false in case of an error.
var isSubnetTaken = func(subnet string) (bool, error) {
	ifAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, fmt.Errorf("failed listing network interface addresses: %w", err)
	}
	for _, ifAddr := range ifAddrs {
		_, lan, err := net.ParseCIDR(ifAddr.String())
		if err != nil {
			return false, fmt.Errorf("failed parsing network interface address %+v: %w", ifAddr, err)
		}
		if lan.Contains(net.ParseIP(subnet)) {
			return true, nil
		}
	}
	return false, nil
}

// IsBuiltinQEMU returns if network is builtin or the legacy value user.
func IsBuiltinQEMU(network string) bool {
	if network == "user" {
		return true
	}
	return network == "builtin"
}

// FreeSubnet will try to find free private network beginning with startSubnet, incrementing it in steps up to number of tries.
func FreeSubnet(startSubnet string, step, tries int) (*Parameters, error) {
	currSubnet := startSubnet
	for try := 0; try < tries; try++ {
		n, err := Inspect(currSubnet)
		if err != nil {
			return nil, err
		}
		subnet := n.IP
		if n.IsPrivate {
			taken, err := isSubnetTaken(subnet)
			if err != nil {
				return nil, err
			}
			if !taken {
				if reservation, err := reserveSubnet(subnet); err == nil {
					n.reservation = reservation
					klog.Infof("using free private subnet %s: %+v", n.CIDR, n)
					return n, nil
				}
				klog.Infof("skipping subnet %s that is reserved: %+v", n.CIDR, n)
			} else {
				klog.Infof("skipping subnet %s that is taken: %+v", n.CIDR, n)
			}
		} else {
			klog.Infof("skipping subnet %s that is not private", n.CIDR)
		}
		prefix, _ := net.ParseIP(n.IP).DefaultMask().Size()
		nextSubnet := net.ParseIP(currSubnet).To4()
		if prefix <= 16 {
			nextSubnet[1] += byte(step)
		} else {
			nextSubnet[2] += byte(step)
		}
		currSubnet = nextSubnet.String()
	}
	return nil, fmt.Errorf("no free private network subnets found with given parameters (start: %q, step: %d, tries: %d)", startSubnet, step, tries)
}

// ParseAddr will try to parse an ip or a cidr address
func ParseAddr(addr string) (net.IP, *net.IPNet, error) {
	ip, network, err := net.ParseCIDR(addr)
	if err != nil {
		ip = net.ParseIP(addr)
		if ip == nil {
			return nil, nil, fmt.Errorf("failed parsing address %s: %w", addr, err)
		}
		err = nil
	}
	return ip, network, err
}

// reserveSubnet returns releaser if subnet was successfully reserved, creating lock for subnet to avoid race condition between multiple minikube instances (especially while testing in parallel).
var reserveSubnet = func(subnet string) (mutex.Releaser, error) {
	spec := lock.PathMutexSpec(subnet)
	spec.Timeout = 1 * time.Millisecond // practically: just check, don't wait
	reservation, err := mutex.Acquire(spec)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}
