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

	"github.com/pmezard/go-difflib/difflib"
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
ExecStart=/var/lib/minikube/binaries/v1.12.0/kubelet --allow-privileged=true --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cadvisor-port=0 --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-domain=cluster.local --config=/var/lib/kubelet/config.yaml --container-runtime=docker --fail-swap-on=false --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.100 --pod-manifest-path=/etc/kubernetes/manifests

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
ExecStart=/var/lib/minikube/binaries/v1.18.2/kubelet --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-domain=cluster.local --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=/var/run/crio/crio.sock --fail-swap-on=false --hostname-override=minikube --image-service-endpoint=/var/run/crio/crio.sock --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.100 --pod-manifest-path=/etc/kubernetes/manifests --runtime-request-timeout=15m

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
ExecStart=/var/lib/minikube/binaries/v1.18.2/kubelet --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-domain=cluster.local --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --fail-swap-on=false --hostname-override=minikube --image-service-endpoint=unix:///run/containerd/containerd.sock --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.100 --pod-manifest-path=/etc/kubernetes/manifests --runtime-request-timeout=15m

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
ExecStart=/var/lib/minikube/binaries/v1.18.2/kubelet --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-domain=cluster.local --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --fail-swap-on=false --hostname-override=minikube --image-service-endpoint=unix:///run/containerd/containerd.sock --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.200 --pod-manifest-path=/etc/kubernetes/manifests --runtime-request-timeout=15m

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
ExecStart=/var/lib/minikube/binaries/v1.18.2/kubelet --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-domain=cluster.local --config=/var/lib/kubelet/config.yaml --container-runtime=docker --fail-swap-on=false --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.1.100 --pod-infra-container-image=docker-proxy-image.io/google_containers/pause:3.2 --pod-manifest-path=/etc/kubernetes/manifests

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

			diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(tc.expected),
				B:        difflib.SplitLines(string(got)),
				FromFile: "Expected",
				ToFile:   "Got",
				Context:  1,
			})
			if err != nil {
				t.Fatalf("diff error: %v\n%s", err, diff)
			}
		})
	}
}
