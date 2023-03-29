/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package images

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/version"
)

func TestKubeadmImages(t *testing.T) {
	tests := []struct {
		version string
		mirror  string
		invalid bool
		want    []string
	}{
		{"invalid", "", true, nil},
		{"v0.0.1", "", true, nil}, // too old
		{"v2.0.0", "", true, nil}, // too new
		{"v1.26.0-rc.0", "", false, []string{
			"registry.k8s.io/kube-apiserver:v1.26.0-rc.0",
			"registry.k8s.io/kube-controller-manager:v1.26.0-rc.0",
			"registry.k8s.io/kube-scheduler:v1.26.0-rc.0",
			"registry.k8s.io/kube-proxy:v1.26.0-rc.0",
			"registry.k8s.io/coredns/coredns:v1.9.3",
			"registry.k8s.io/etcd:3.5.5-0",
			"registry.k8s.io/pause:3.9",
			"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		}},
		{"v1.25.4", "", false, []string{
			"registry.k8s.io/kube-apiserver:v1.25.4",
			"registry.k8s.io/kube-controller-manager:v1.25.4",
			"registry.k8s.io/kube-scheduler:v1.25.4",
			"registry.k8s.io/kube-proxy:v1.25.4",
			"registry.k8s.io/coredns/coredns:v1.9.3",
			"registry.k8s.io/etcd:3.5.5-0",
			"registry.k8s.io/pause:3.8",
			"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		}},
		{"v1.25.0", "", false, []string{
			"registry.k8s.io/kube-proxy:v1.25.0",
			"registry.k8s.io/kube-scheduler:v1.25.0",
			"registry.k8s.io/kube-controller-manager:v1.25.0",
			"registry.k8s.io/kube-apiserver:v1.25.0",
			"registry.k8s.io/coredns/coredns:v1.9.3",
			"registry.k8s.io/etcd:3.5.4-0",
			"registry.k8s.io/pause:3.8",
			"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		}},
		{"v1.25.0", "mirror.k8s.io", false, []string{
			"mirror.k8s.io/kube-proxy:v1.25.0",
			"mirror.k8s.io/kube-scheduler:v1.25.0",
			"mirror.k8s.io/kube-controller-manager:v1.25.0",
			"mirror.k8s.io/kube-apiserver:v1.25.0",
			"mirror.k8s.io/coredns/coredns:v1.9.3",
			"mirror.k8s.io/etcd:3.5.4-0",
			"mirror.k8s.io/pause:3.8",
			"mirror.k8s.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		}},
		{"v1.24.0", "", false, []string{
			"registry.k8s.io/kube-proxy:v1.24.0",
			"registry.k8s.io/kube-scheduler:v1.24.0",
			"registry.k8s.io/kube-controller-manager:v1.24.0",
			"registry.k8s.io/kube-apiserver:v1.24.0",
			"registry.k8s.io/coredns/coredns:v1.8.6",
			"registry.k8s.io/etcd:3.5.3-0",
			"registry.k8s.io/pause:3.7",
			"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		}},
		{"v1.23.0", "", false, []string{
			"registry.k8s.io/kube-proxy:v1.23.0",
			"registry.k8s.io/kube-scheduler:v1.23.0",
			"registry.k8s.io/kube-controller-manager:v1.23.0",
			"registry.k8s.io/kube-apiserver:v1.23.0",
			"registry.k8s.io/coredns/coredns:v1.8.6",
			"registry.k8s.io/etcd:3.5.1-0",
			"registry.k8s.io/pause:3.6",
			"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		}},
		{"v1.22.0", "", false, []string{
			"registry.k8s.io/kube-proxy:v1.22.0",
			"registry.k8s.io/kube-scheduler:v1.22.0",
			"registry.k8s.io/kube-controller-manager:v1.22.0",
			"registry.k8s.io/kube-apiserver:v1.22.0",
			"registry.k8s.io/coredns/coredns:v1.8.4",
			"registry.k8s.io/etcd:3.5.0-0",
			"registry.k8s.io/pause:3.5",
			"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		}},
		{"v1.16.0", "", false, []string{
			"registry.k8s.io/kube-proxy:v1.16.0",
			"registry.k8s.io/kube-scheduler:v1.16.0",
			"registry.k8s.io/kube-controller-manager:v1.16.0",
			"registry.k8s.io/kube-apiserver:v1.16.0",
			"registry.k8s.io/coredns:1.6.2",
			"registry.k8s.io/etcd:3.3.15-0",
			"registry.k8s.io/pause:3.1",
			"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		}},
		{"v1.11.0", "", true, nil},
		{"v1.10.0", "", true, nil},
	}
	for _, tc := range tests {
		got, err := Kubeadm(tc.mirror, tc.version)
		if err == nil && tc.invalid {
			t.Fatalf("expected err (%s): %v", tc.version, got)
		}
		if err != nil && !tc.invalid {
			t.Fatalf("unexpected err (%s): %v", tc.version, err)
		}
		sort.Strings(got)
		sort.Strings(tc.want)
		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("%s images mismatch (-want +got):\n%s", tc.version, diff)
		}
	}
}
