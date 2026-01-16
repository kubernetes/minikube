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

package util

import (
	"net"

	"github.com/pkg/errors"
)

// DefaultAdmissionControllers are admission controllers we default to
var DefaultAdmissionControllers = []string{
	"NamespaceLifecycle",
	"LimitRanger",
	"ServiceAccount",
	"DefaultStorageClass",
	"DefaultTolerationSeconds",
	"NodeRestriction",
	"MutatingAdmissionWebhook",
	"ValidatingAdmissionWebhook",
	"ResourceQuota",
}

// ServiceClusterIP returns the first usable IP of the Service CIDR (network + 1) for either IPv4 or IPv6.
func ServiceClusterIP(serviceCIDR string) (net.IP, error) {
	ip, _, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "parsing default service cidr")
	}
	base := normalizeIP(ip)
	if base == nil {
		return nil, errors.Errorf("invalid serviceCIDR base IP: %q", serviceCIDR)
	}
	out, ok := addToIP(base, 1)
	if !ok {
		return nil, errors.Errorf("serviceCIDR %q has no usable service IPs", serviceCIDR)
	}
	return out, nil
}

func DNSIP(serviceCIDR string) (net.IP, error) {
	ip, _, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "parsing default service cidr")
	}
	base := normalizeIP(ip)
	if base == nil {
		return nil, errors.Errorf("invalid serviceCIDR base IP: %q", serviceCIDR)
	}
	out, ok := addToIP(base, 10)
	if !ok {
		return nil, errors.Errorf("serviceCIDR %q too small for DNS IP allocation", serviceCIDR)
	}
	return out, nil
}

// AlternateDNS returns a list of alternate names for a domain
func AlternateDNS(domain string) []string {
	return []string{"kubernetes.default.svc." + domain, "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}
}

// normalizeIP returns a 4-byte slice for v4 or 16-byte slice for v6.
func normalizeIP(ip net.IP) net.IP {
	if ip == nil {
		return nil
	}
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	if v6 := ip.To16(); v6 != nil {
		return v6
	}
	return nil
}

// addToIP returns ip + n, preserving length (v4/v6) with carry.
// ok=false if ip is not v4/v6 length, or if addition overflows the address space.
func addToIP(ip net.IP, n uint64) (out net.IP, ok bool) {
	if ip == nil {
		return nil, false
	}
	if len(ip) != net.IPv4len && len(ip) != net.IPv6len {
		return nil, false
	}
	out = make(net.IP, len(ip))
	copy(out, ip)
	i := len(out) - 1
	for n > 0 && i >= 0 {
		sum := uint64(out[i]) + (n & 0xff)
		out[i] = byte(sum & 0xff)
		n = (n >> 8) + (sum >> 8)
		i--
	}

	if n > 0 {
		return nil, false
	}
	return out, true
}
