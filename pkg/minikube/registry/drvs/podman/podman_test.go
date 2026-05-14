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

package podman

import (
	"testing"

	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/minikube/config"
)

func TestConfigureDNSSearchEnv(t *testing.T) {
	tests := []struct {
		name      string
		dnsSearch []string
		wantEnv   string
	}{
		{
			name:      "no dns search",
			dnsSearch: nil,
			wantEnv:   "",
		},
		{
			name:      "single domain",
			dnsSearch: []string{"corp.example.com"},
			wantEnv:   "corp.example.com",
		},
		{
			name:      "multiple domains",
			dnsSearch: []string{"corp.example.com", "eng.example.com"},
			wantEnv:   "corp.example.com eng.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := config.ClusterConfig{
				Name:         "test",
				KicBaseImage: "gcr.io/k8s-minikube/kicbase:test",
				DNSSearch:    tt.dnsSearch,
				Nodes: []config.Node{
					{
						Name:         "test",
						Port:         8443,
						ControlPlane: true,
					},
				},
				KubernetesConfig: config.KubernetesConfig{
					ContainerRuntime: "containerd",
				},
			}
			n := cc.Nodes[0]

			result, err := configure(cc, n)
			if err != nil {
				t.Fatalf("configure() returned error: %v", err)
			}

			d, ok := result.(*kic.Driver)
			if !ok {
				t.Fatalf("configure() returned %T, want *kic.Driver", result)
			}

			gotEnv := d.NodeConfig.Envs["KIND_DNS_SEARCH"]
			if gotEnv != tt.wantEnv {
				t.Errorf("KIND_DNS_SEARCH = %q, want %q", gotEnv, tt.wantEnv)
			}
		})
	}
}
