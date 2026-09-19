/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package kic

import (
	"net"
	"testing"
)

func TestContainerNetwork(t *testing.T) {
	tests := []struct {
		name        string
		gateway     net.IP
		networkName string
		staticIP    string
		machineName string
		wantNetwork string
		wantIP      string
		wantErr     bool
	}{
		{
			// https://github.com/kubernetes/minikube/issues/19284
			name:        "static IP on existing network without gateway",
			gateway:     nil,
			networkName: "test_net",
			staticIP:    "192.168.150.20",
			machineName: "minikube",
			wantNetwork: "test_net",
			wantIP:      "192.168.150.20",
		},
		{
			name:        "static IP on network with gateway",
			gateway:     net.ParseIP("192.168.150.1"),
			networkName: "test_net",
			staticIP:    "192.168.150.20",
			machineName: "minikube",
			wantNetwork: "test_net",
			wantIP:      "192.168.150.20",
		},
		{
			name:        "calculated IP for first node",
			gateway:     net.ParseIP("192.168.49.1"),
			networkName: "minikube",
			machineName: "minikube",
			wantNetwork: "minikube",
			wantIP:      "192.168.49.2",
		},
		{
			name:        "calculated IP for second node",
			gateway:     net.ParseIP("192.168.49.1"),
			networkName: "minikube",
			machineName: "minikube-m02",
			wantNetwork: "minikube",
			wantIP:      "192.168.49.3",
		},
		{
			name:        "default network without gateway",
			gateway:     nil,
			networkName: "bridge",
			machineName: "minikube",
		},
		{
			name:        "too many machines",
			gateway:     net.ParseIP("192.168.49.253"),
			networkName: "minikube",
			machineName: "minikube",
			wantErr:     true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			network, ip, err := containerNetwork(tc.gateway, tc.networkName, tc.staticIP, tc.machineName)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got network %q and IP %q", network, ip)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if network != tc.wantNetwork {
				t.Errorf("expected network %q, got %q", tc.wantNetwork, network)
			}
			if ip != tc.wantIP {
				t.Errorf("expected IP %q, got %q", tc.wantIP, ip)
			}
		})
	}
}
