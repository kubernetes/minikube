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

// These constants are used by both minikube
const (
	APIServerPort            = 8443
	DefaultMinikubeDirectory = "/var/lib/minikube"
	DefaultCertPath          = DefaultMinikubeDirectory + "/certs/"
	DefaultKubeConfigPath    = DefaultMinikubeDirectory + "/kubeconfig"
	DefaultDNSDomain         = "cluster.local"
	DefaultServiceCIDR       = "10.96.0.0/12"
)

var DefaultAdmissionControllers = []string{
	"Initializers",
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

// GetServiceClusterIP returns the first IP of the ServiceCIDR
func GetServiceClusterIP(serviceCIDR string) (net.IP, error) {
	ip, _, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "parsing default service cidr")
	}
	ip = ip.To4()
	ip[3]++
	return ip, nil
}

// GetDNSIP returns x.x.x.10 of the service CIDR
func GetDNSIP(serviceCIDR string) (net.IP, error) {
	ip, _, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "parsing default service cidr")
	}
	ip = ip.To4()
	ip[3] = 10
	return ip, nil
}

func GetAlternateDNS(domain string) []string {
	return []string{"kubernetes.default.svc." + domain, "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}
}
