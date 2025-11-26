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
	"fmt"
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
// - If serviceCIDR is empty, DefaultServiceCIDR is used.
// - If multiple CIDRs are provided (comma-separated), IPv4 is preferred for
//   backward compatibility; otherwise the first non-empty CIDR is used.
// - For IPv4, we return the first usable IP after the network address.
// - For IPv6, we return the first IP in the subnet (or +1 on the last byte –
//   the exact value is not important as long as it's inside the CIDR).
func ServiceClusterIP(serviceCIDR string) (net.IP, error) {
    if serviceCIDR == "" {
        serviceCIDR = mkconstants.DefaultServiceCIDR
    }

    cidr := pickPrimaryServiceCIDR(serviceCIDR)
    cidr = strings.TrimSpace(cidr)
    if cidr == "" {
        return nil, fmt.Errorf("empty service CIDR")
    }

    _, ipNet, err := net.ParseCIDR(cidr)
    if err != nil {
        return nil, err
    }

    // Copy to avoid mutating ipNet.IP
    ip := make(net.IP, len(ipNet.IP))
    copy(ip, ipNet.IP)

    // IPv4: bump last octet (avoid the network address itself).
    if v4 := ip.To4(); v4 != nil {
        v4[3]++
        return v4, nil
    }

    // IPv6: just bump the last byte – stays within the subnet for normal masks.
    if len(ip) == 0 {
        return nil, fmt.Errorf("parsed service CIDR %q has empty IP", cidr)
    }
    ip[len(ip)-1]++
    return ip, nil
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
	base, _, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "parsing default service cidr")
	}
	ip := normalizeIP(base)
	return addToIP(ip, 10), nil
}

// AlternateDNS returns a list of alternate names for a domain
func AlternateDNS(domain string) []string {
	return []string{"kubernetes.default.svc." + domain, "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}
}

// normalizeIP returns a 4-byte slice for v4 or 16-byte slice for v6.
func normalizeIP(ip net.IP) net.IP {
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip.To16()
}

// addToIP returns ip + n, preserving length (v4/v6) with carry.
func addToIP(ip net.IP, n uint64) net.IP {
	out := make(net.IP, len(ip))
	copy(out, ip)
	i := len(out) - 1
	for n > 0 && i >= 0 {
		sum := uint64(out[i]) + (n & 0xff)
		out[i] = byte(sum & 0xff)
		n = (n >> 8) + (sum >> 8)
		i--
	}
	return out
}
