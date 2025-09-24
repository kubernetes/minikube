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
       base, _, err := net.ParseCIDR(serviceCIDR)
       if err != nil {
               return nil, errors.Wrap(err, "parsing default service cidr")
       }
       ip := normalizeIP(base)
       return addToIP(ip, 1), nil
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
