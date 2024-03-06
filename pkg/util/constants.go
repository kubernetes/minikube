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

// ServiceClusterIP returns the first IP of the ServiceCIDR
func ServiceClusterIP(serviceCIDR string) (net.IP, error) {
	ip, _, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "parsing default service cidr")
	}
	ip = ip.To4()
	ip[3]++
	return ip, nil
}

// DNSIP returns x.x.x.10 of the service CIDR
func DNSIP(serviceCIDR string) (net.IP, error) {
	ip, _, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "parsing default service cidr")
	}
	ip = ip.To4()
	ip[3] = 10
	return ip, nil
}

// AlternateDNS returns a list of alternate names for a domain
func AlternateDNS(domain string) []string {
	return []string{"kubernetes.default.svc." + domain, "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}
}
