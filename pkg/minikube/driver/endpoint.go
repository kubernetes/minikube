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
	"net"

	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// ControlPaneEndpoint returns the location where callers can reach this cluster
func ControlPaneEndpoint(cc *config.ClusterConfig, cp *config.Node, driverName string) (string, net.IP, int, error) {
	if NeedsPortForward(driverName) {
		port, err := oci.ForwardedPort(cc.Driver, cc.Name, cp.Port)
		hostname := oci.DefaultBindIPV4
		ip := net.ParseIP(hostname)

		// https://github.com/kubernetes/minikube/issues/3878
		if cc.KubernetesConfig.APIServerName != constants.APIServerName {
			hostname = cc.KubernetesConfig.APIServerName
		}
		return hostname, ip, port, err
	}

	// https://github.com/kubernetes/minikube/issues/3878
	hostname := cp.IP
	if cc.KubernetesConfig.APIServerName != constants.APIServerName {
		hostname = cc.KubernetesConfig.APIServerName
	}
	return hostname, net.ParseIP(cp.IP), cp.Port, nil
}
