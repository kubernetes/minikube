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
			"k8s.gcr.io/kube-proxy" + archTag(false) + "v1.17.0",
			"k8s.gcr.io/kube-scheduler" + archTag(false) + "v1.17.0",
			"k8s.gcr.io/kube-controller-manager" + archTag(false) + "v1.17.0",
			"k8s.gcr.io/kube-apiserver" + archTag(false) + "v1.17.0",
			"k8s.gcr.io/coredns:1.6.5",
			"k8s.gcr.io/etcd" + archTag(false) + "3.4.3-0",
			"k8s.gcr.io/pause" + archTag(false) + "3.1",
			"gcr.io/k8s-minikube/storage-provisioner" + archTag(false) + "v1.8.1",
			"kubernetesui/dashboard:v2.0.0-beta8",
			"kubernetesui/metrics-scraper:v1.0.2",
		}},
		{"v1.16.1", "mirror.k8s.io", []string{
			"mirror.k8s.io/kube-proxy" + archTag(false) + "v1.16.1",
			"mirror.k8s.io/kube-scheduler" + archTag(false) + "v1.16.1",
			"mirror.k8s.io/kube-controller-manager" + archTag(false) + "v1.16.1",
			"mirror.k8s.io/kube-apiserver" + archTag(false) + "v1.16.1",
			"mirror.k8s.io/coredns:1.6.2",
			"mirror.k8s.io/etcd" + archTag(false) + "3.3.15-0",
			"mirror.k8s.io/pause" + archTag(false) + "3.1",
			"mirror.k8s.io/storage-provisioner" + archTag(false) + "v1.8.1",
			"mirror.k8s.io/dashboard:v2.0.0-beta8",
			"mirror.k8s.io/metrics-scraper:v1.0.2",
		}},
		{"v1.15.0", "", []string{
			"k8s.gcr.io/kube-proxy" + archTag(false) + "v1.15.0",
			"k8s.gcr.io/kube-scheduler" + archTag(false) + "v1.15.0",
			"k8s.gcr.io/kube-controller-manager" + archTag(false) + "v1.15.0",
			"k8s.gcr.io/kube-apiserver" + archTag(false) + "v1.15.0",
			"k8s.gcr.io/coredns:1.3.1",
			"k8s.gcr.io/etcd" + archTag(false) + "3.3.10",
			"k8s.gcr.io/pause" + archTag(false) + "3.1",
			"gcr.io/k8s-minikube/storage-provisioner" + archTag(false) + "v1.8.1",
			"kubernetesui/dashboard:v2.0.0-beta8",
			"kubernetesui/metrics-scraper:v1.0.2",
		}},
		{"v1.14.0", "", []string{
			"k8s.gcr.io/kube-proxy" + archTag(false) + "v1.14.0",
			"k8s.gcr.io/kube-scheduler" + archTag(false) + "v1.14.0",
			"k8s.gcr.io/kube-controller-manager" + archTag(false) + "v1.14.0",
			"k8s.gcr.io/kube-apiserver" + archTag(false) + "v1.14.0",
			"k8s.gcr.io/coredns:1.3.1",
			"k8s.gcr.io/etcd" + archTag(false) + "3.3.10",
			"k8s.gcr.io/pause" + archTag(false) + "3.1",
			"gcr.io/k8s-minikube/storage-provisioner" + archTag(false) + "v1.8.1",
			"kubernetesui/dashboard:v2.0.0-beta8",
			"kubernetesui/metrics-scraper:v1.0.2",
		}},
		{"v1.13.0", "", []string{
			"k8s.gcr.io/kube-proxy" + archTag(false) + "v1.13.0",
			"k8s.gcr.io/kube-scheduler" + archTag(false) + "v1.13.0",
			"k8s.gcr.io/kube-controller-manager" + archTag(false) + "v1.13.0",
			"k8s.gcr.io/kube-apiserver" + archTag(false) + "v1.13.0",
			"k8s.gcr.io/coredns:1.2.6",
			"k8s.gcr.io/etcd" + archTag(false) + "3.2.24",
			"k8s.gcr.io/pause" + archTag(false) + "3.1",
			"gcr.io/k8s-minikube/storage-provisioner" + archTag(false) + "v1.8.1",
			"kubernetesui/dashboard:v2.0.0-beta8",
			"kubernetesui/metrics-scraper:v1.0.2",
		}},
		{"v1.12.0", "", []string{
			"k8s.gcr.io/kube-proxy" + archTag(false) + "v1.12.0",
			"k8s.gcr.io/kube-scheduler" + archTag(false) + "v1.12.0",
			"k8s.gcr.io/kube-controller-manager" + archTag(false) + "v1.12.0",
			"k8s.gcr.io/kube-apiserver" + archTag(false) + "v1.12.0",
			"k8s.gcr.io/coredns:1.2.2",
			"k8s.gcr.io/etcd" + archTag(false) + "3.2.24",
			"k8s.gcr.io/pause" + archTag(false) + "3.1",
			"gcr.io/k8s-minikube/storage-provisioner" + archTag(false) + "v1.8.1",
			"kubernetesui/dashboard:v2.0.0-beta8",
			"kubernetesui/metrics-scraper:v1.0.2",
		}},
		{"v1.11.10", "", []string{
			"k8s.gcr.io/kube-proxy" + archTag(true) + "v1.11.10",
			"k8s.gcr.io/kube-scheduler" + archTag(true) + "v1.11.10",
			"k8s.gcr.io/kube-controller-manager" + archTag(true) + "v1.11.10",
			"k8s.gcr.io/kube-apiserver" + archTag(true) + "v1.11.10",
			"k8s.gcr.io/coredns:1.1.3",
			"k8s.gcr.io/etcd" + archTag(true) + "3.2.18",
			"k8s.gcr.io/pause" + archTag(false) + "3.1",
			"gcr.io/k8s-minikube/storage-provisioner" + archTag(false) + "v1.8.1",
			"kubernetesui/dashboard:v2.0.0-beta8",
			"kubernetesui/metrics-scraper:v1.0.2",
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
