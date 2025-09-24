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

package driver

import (
	"fmt"
	"net"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/network"
)

func ControlPlaneEndpoint(cc *config.ClusterConfig, cp *config.Node, driverName string) (string, net.IP, int, error) {
	if NeedsPortForward(driverName) {
		port, err := oci.ForwardedPort(cc.Driver, cc.Name, cp.Port)
		if err != nil {
			klog.Warningf("failed to get forwarded control plane port %v", err)
		}


		// Start with daemon host (docker/podman), tweak for IPv6, then honor APIServerName override.
                host := oci.DaemonHost(driverName)
                // If the cluster/node IP is IPv6 and daemon host is localhost on IPv4,
                // force IPv6 loopback so we hit the port that’s actually listening.
                if strings.Contains(cp.IP, ":") && (host == "127.0.0.1" || host == "localhost") {
                        host = "::1"
                }
                if cc.KubernetesConfig.APIServerName != constants.APIServerName {
                        host = cc.KubernetesConfig.APIServerName
                }

		// Resolve final host -> IPs. Allow literal IPv4/IPv6 without DNS.
		var ips []net.IP
		if ip := net.ParseIP(host); ip != nil {
			ips = []net.IP{ip}
		} else {
			ips, err = net.LookupIP(host)
			if err != nil || len(ips) == 0 {
				return host, nil, port, fmt.Errorf("failed to lookup ip for %q", host)
			}
		}


		return host, ips[0], port, nil
	}

	if IsQEMU(driverName) && network.IsBuiltinQEMU(cc.Network) {
                if strings.Contains(cp.IP, ":") {
                        return "::1", net.IPv6loopback, cc.APIServerPort, nil
                }
                return "127.0.0.1", net.IPv4(127, 0, 0, 1), cc.APIServerPort, nil
	}

	// Default: use the node IP (literal or resolvable name)
	host := cp.IP
	if cc.KubernetesConfig.APIServerName != constants.APIServerName {
		host = cc.KubernetesConfig.APIServerName
	}

	var ips []net.IP
	if ip := net.ParseIP(cp.IP); ip != nil {
		ips = []net.IP{ip}
	} else {
		var err error
		ips, err = net.LookupIP(cp.IP)
		if err != nil || len(ips) == 0 {
			return host, nil, cp.Port, fmt.Errorf("failed to lookup ip for %q", cp.IP)
		}
	}

	return host, ips[0], cp.Port, nil
}

// AutoPauseProxyEndpoint returns the endpoint for the auto-pause (reverse proxy to api-sever)
func AutoPauseProxyEndpoint(cc *config.ClusterConfig, cp *config.Node, driverName string) (string, net.IP, int, error) {
	cp.Port = constants.AutoPauseProxyPort
	return ControlPlaneEndpoint(cc, cp, driverName)
}
