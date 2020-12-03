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
)

func TestKubeadmImages(t *testing.T) {
	tests := []struct {
		version string
		mirror  string
		want    []string
	}{
		{"v1.17.0", "", []string{
			"k8s.gcr.io/kube-proxy:v1.17.0",
			"k8s.gcr.io/kube-scheduler:v1.17.0",
			"k8s.gcr.io/kube-controller-manager:v1.17.0",
			"k8s.gcr.io/kube-apiserver:v1.17.0",
			"k8s.gcr.io/coredns:1.6.5",
			"k8s.gcr.io/etcd:3.4.3-0",
			"k8s.gcr.io/pause:3.1",
			"gcr.io/k8s-minikube/storage-provisioner:v4",
			"docker.io/kubernetesui/dashboard:v2.0.3",
			"docker.io/kubernetesui/metrics-scraper:v1.0.4",
		}},
		{"v1.16.1", "mirror.k8s.io", []string{
			"mirror.k8s.io/kube-proxy:v1.16.1",
			"mirror.k8s.io/kube-scheduler:v1.16.1",
			"mirror.k8s.io/kube-controller-manager:v1.16.1",
			"mirror.k8s.io/kube-apiserver:v1.16.1",
			"mirror.k8s.io/coredns:1.6.2",
			"mirror.k8s.io/etcd:3.3.15-0",
			"mirror.k8s.io/pause:3.1",
			"mirror.k8s.io/storage-provisioner:v4",
			"mirror.k8s.io/dashboard:v2.0.3",
			"mirror.k8s.io/metrics-scraper:v1.0.4",
		}},
		{"v1.15.0", "", []string{
			"k8s.gcr.io/kube-proxy:v1.15.0",
			"k8s.gcr.io/kube-scheduler:v1.15.0",
			"k8s.gcr.io/kube-controller-manager:v1.15.0",
			"k8s.gcr.io/kube-apiserver:v1.15.0",
			"k8s.gcr.io/coredns:1.3.1",
			"k8s.gcr.io/etcd:3.3.10",
			"k8s.gcr.io/pause:3.1",
			"gcr.io/k8s-minikube/storage-provisioner:v4",
			"docker.io/kubernetesui/dashboard:v2.0.3",
			"docker.io/kubernetesui/metrics-scraper:v1.0.4",
		}},
		{"v1.14.0", "", []string{
			"k8s.gcr.io/kube-proxy:v1.14.0",
			"k8s.gcr.io/kube-scheduler:v1.14.0",
			"k8s.gcr.io/kube-controller-manager:v1.14.0",
			"k8s.gcr.io/kube-apiserver:v1.14.0",
			"k8s.gcr.io/coredns:1.3.1",
			"k8s.gcr.io/etcd:3.3.10",
			"k8s.gcr.io/pause:3.1",
			"gcr.io/k8s-minikube/storage-provisioner:v4",
			"docker.io/kubernetesui/dashboard:v2.0.3",
			"docker.io/kubernetesui/metrics-scraper:v1.0.4",
		}},
		{"v1.13.0", "", []string{
			"k8s.gcr.io/kube-proxy:v1.13.0",
			"k8s.gcr.io/kube-scheduler:v1.13.0",
			"k8s.gcr.io/kube-controller-manager:v1.13.0",
			"k8s.gcr.io/kube-apiserver:v1.13.0",
			"k8s.gcr.io/coredns:1.2.6",
			"k8s.gcr.io/etcd:3.2.24",
			"k8s.gcr.io/pause:3.1",
			"gcr.io/k8s-minikube/storage-provisioner:v4",
			"docker.io/kubernetesui/dashboard:v2.0.3",
			"docker.io/kubernetesui/metrics-scraper:v1.0.4",
		}},
		{"v1.12.0", "", []string{
			"k8s.gcr.io/kube-proxy:v1.12.0",
			"k8s.gcr.io/kube-scheduler:v1.12.0",
			"k8s.gcr.io/kube-controller-manager:v1.12.0",
			"k8s.gcr.io/kube-apiserver:v1.12.0",
			"k8s.gcr.io/coredns:1.2.2",
			"k8s.gcr.io/etcd:3.2.24",
			"k8s.gcr.io/pause:3.1",
			"gcr.io/k8s-minikube/storage-provisioner:v4",
			"docker.io/kubernetesui/dashboard:v2.0.3",
			"docker.io/kubernetesui/metrics-scraper:v1.0.4",
		}},
	}
	for _, tc := range tests {
		got, err := Kubeadm(tc.mirror, tc.version)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		sort.Strings(got)
		sort.Strings(tc.want)
		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("%s images mismatch (-want +got):\n%s", tc.version, diff)
		}
	}
}
