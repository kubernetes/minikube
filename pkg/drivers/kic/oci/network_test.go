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

package oci

import "testing"

func TestNetworkIPs(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		network string
		ipv4    string
		ipv6    string
		wantErr bool
	}{
		{
			name:    "single network",
			lines:   []string{"minikube,192.168.49.2,"},
			network: "minikube",
			ipv4:    "192.168.49.2",
		},
		{
			name:    "extra network sorted first",
			lines:   []string{"bridge,172.17.0.2,", "minikube,192.168.49.2,fd00::2"},
			network: "minikube",
			ipv4:    "192.168.49.2",
			ipv6:    "fd00::2",
		},
		{
			name:    "single network is not the named one",
			lines:   []string{"bridge,172.17.0.2,", ""},
			network: "minikube",
			ipv4:    "172.17.0.2",
		},
		{
			name:    "named network is not attached",
			lines:   []string{"bridge,172.17.0.2,", "extra,172.20.0.2,"},
			network: "minikube",
			wantErr: true,
		},
		{
			name:    "malformed line",
			lines:   []string{"minikube,192.168.49.2"},
			network: "minikube",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ipv4, ipv6, err := networkIPs(tc.lines, tc.network)
			if (err != nil) != tc.wantErr {
				t.Fatalf("networkIPs(%q, %q) error = %v, wantErr %v", tc.lines, tc.network, err, tc.wantErr)
			}
			if ipv4 != tc.ipv4 || ipv6 != tc.ipv6 {
				t.Errorf("networkIPs(%q, %q) = %q, %q, want %q, %q", tc.lines, tc.network, ipv4, ipv6, tc.ipv4, tc.ipv6)
			}
		})
	}
}
