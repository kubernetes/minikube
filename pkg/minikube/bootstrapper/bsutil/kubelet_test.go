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

// Package bsutil will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

func TestGenerateKubeletConfig(t *testing.T) {
	tests := []struct {
		description string
		cfg         config.ClusterConfig
		expected    string
		shouldErr   bool
	}{
		{
			description: "old docker",
			cfg: config.ClusterConfig{
				Name: "minikube",
				KubernetesConfig: config.KubernetesConfig{
					KubernetesVersion: constants.OldestKubernetesVersion,
					ContainerRuntime:  "docker",
				},
				Nodes: []config.Node{
					{
						IP:           "192.168.1.100",
						Name:         "minikube",
						ControlPlane: true,
					},
				},
			},
			expected: `[Unit]
Wants=docker.socket

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.28.0/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=docker --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.100

[Install]
`,
		},
		{
			description: "newest cri runtime",
			cfg: config.ClusterConfig{
				Name: "minikube",
				KubernetesConfig: config.KubernetesConfig{
					KubernetesVersion: constants.NewestKubernetesVersion,
					ContainerRuntime:  "cri-o",
				},
				Nodes: []config.Node{
					{
						IP:           "192.168.1.100",
						Name:         "minikube",
						ControlPlane: true,
					},
				},
			},
			expected: `[Unit]
Wants=crio.service

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.35.0/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.100

[Install]
`,
		},
		{
			description: "default containerd runtime",
			cfg: config.ClusterConfig{
				Name: "minikube",
				KubernetesConfig: config.KubernetesConfig{
					KubernetesVersion: constants.DefaultKubernetesVersion,
					ContainerRuntime:  "containerd",
				},
				Nodes: []config.Node{
					{
						IP:           "192.168.1.100",
						Name:         "minikube",
						ControlPlane: true,
					},
				},
			},
			expected: `[Unit]
Wants=containerd.service

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.35.0/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.100

[Install]
`,
		},
		{
			description: "default containerd runtime with IP override",
			cfg: config.ClusterConfig{
				Name: "minikube",
				KubernetesConfig: config.KubernetesConfig{
					KubernetesVersion: constants.DefaultKubernetesVersion,
					ContainerRuntime:  "containerd",
					ExtraOptions: config.ExtraOptionSlice{
						config.ExtraOption{
							Component: Kubelet,
							Key:       "node-ip",
							Value:     "192.168.1.200",
						},
					},
				},
				Nodes: []config.Node{
					{
						IP:           "192.168.1.100",
						Name:         "minikube",
						ControlPlane: true,
					},
				},
			},
			expected: `[Unit]
Wants=containerd.service

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.35.0/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.200

[Install]
`,
		},
		{
			description: "docker with custom image repository",
			cfg: config.ClusterConfig{
				Name: "minikube",
				KubernetesConfig: config.KubernetesConfig{
					KubernetesVersion: constants.DefaultKubernetesVersion,
					ContainerRuntime:  "docker",
					ImageRepository:   "docker-proxy-image.io/google_containers",
				},
				Nodes: []config.Node{
					{
						IP:           "192.168.1.100",
						Name:         "minikube",
						ControlPlane: true,
					},
				},
			},
			expected: `[Unit]
Wants=docker.socket

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.35.0/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=docker --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.100

[Install]
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			runtime, err := cruntime.New(cruntime.Config{Type: tc.cfg.KubernetesConfig.ContainerRuntime})
			if err != nil {
				t.Fatalf("runtime: %v", err)
			}

			got, err := NewKubeletConfig(tc.cfg, tc.cfg.Nodes[0], runtime)
			if err != nil && !tc.shouldErr {
				t.Errorf("got unexpected error generating config: %v", err)
				return
			}
			if err == nil && tc.shouldErr {
				t.Errorf("expected error but got none, config: %s", got)
				return
			}

			if diff := cmp.Diff(tc.expected, string(got)); diff != "" {
				t.Errorf("GenerateKubeletConfig mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
