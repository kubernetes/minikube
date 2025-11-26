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
	"strings"

	"github.com/pkg/errors"
	mkconstants "k8s.io/minikube/pkg/minikube/constants"
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

// ServiceClusterIP returns a "well-known" Service IP from the Service CIDR.
//
//   - If serviceCIDR is empty, DefaultServiceCIDR is used.
//   - If multiple CIDRs are provided (comma-separated), IPv4 is preferred for
//     backward compatibility; otherwise the first non-empty CIDR is used.
//   - For both IPv4 and IPv6, we return the first usable IP after the network address.
func ServiceClusterIP(serviceCIDR string) (net.IP, error) {
	if serviceCIDR == "" {
		serviceCIDR = mkconstants.DefaultServiceCIDR
	}

	cidr := strings.TrimSpace(pickPrimaryServiceCIDR(serviceCIDR))
	if cidr == "" {
		return nil, errors.Errorf("empty service CIDR")
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.Wrap(err, "parsing service CIDR")
	}

	base := normalizeIP(ipNet.IP)
	if base == nil {
		return nil, errors.Errorf("parsed service CIDR %q has invalid base IP", cidr)
	}

	out, ok := addToIP(base, 1)
	if !ok {
		return nil, errors.Errorf("service CIDR %q has no usable service IPs", cidr)
	}
	return out, nil
}

// pickPrimaryServiceCIDR chooses which CIDR to use when a comma-separated list
// is provided. It prefers IPv4 when present, otherwise the first non-empty part.
func pickPrimaryServiceCIDR(serviceCIDR string) string {
	parts := strings.Split(serviceCIDR, ",")
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0])
	}

	var firstNonEmpty string
	var firstV4 string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if firstNonEmpty == "" {
			firstNonEmpty = p
		}
		if strings.Contains(p, ".") && firstV4 == "" {
			firstV4 = p
		}
	}

	if firstV4 != "" {
		return firstV4
	}
	return firstNonEmpty
}

func DNSIP(serviceCIDR string) (net.IP, error) {
	if serviceCIDR == "" {
		serviceCIDR = mkconstants.DefaultServiceCIDR
	}

	cidr := strings.TrimSpace(pickPrimaryServiceCIDR(serviceCIDR))
	if cidr == "" {
		return nil, errors.Errorf("empty service CIDR")
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.Wrap(err, "parsing service CIDR")
	}

	base := normalizeIP(ipNet.IP)
	if base == nil {
		return nil, errors.Errorf("parsed service CIDR %q has invalid base IP", cidr)
	}

	out, ok := addToIP(base, 10)
	if !ok {
		return nil, errors.Errorf("service CIDR %q too small for DNS IP allocation", cidr)
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
