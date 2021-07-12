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
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/version"
)

func TestEssentials(t *testing.T) {
	var testCases = []struct {
		version string
		images  []string
	}{
		{"v1.18.0", strings.Split(strings.Trim(`
k8s.gcr.io/kube-apiserver:v1.18.0
k8s.gcr.io/kube-controller-manager:v1.18.0
k8s.gcr.io/kube-scheduler:v1.18.0
k8s.gcr.io/kube-proxy:v1.18.0
k8s.gcr.io/pause:3.2
k8s.gcr.io/etcd:3.4.3-0
k8s.gcr.io/coredns:1.6.7
`, "\n"), "\n")},
		{"v1.19.0", strings.Split(strings.Trim(`
k8s.gcr.io/kube-apiserver:v1.19.0
k8s.gcr.io/kube-controller-manager:v1.19.0
k8s.gcr.io/kube-scheduler:v1.19.0
k8s.gcr.io/kube-proxy:v1.19.0
k8s.gcr.io/pause:3.2
k8s.gcr.io/etcd:3.4.9-1
k8s.gcr.io/coredns:1.7.0
`, "\n"), "\n")},
		{"v1.20.0", strings.Split(strings.Trim(`
k8s.gcr.io/kube-apiserver:v1.20.0
k8s.gcr.io/kube-controller-manager:v1.20.0
k8s.gcr.io/kube-scheduler:v1.20.0
k8s.gcr.io/kube-proxy:v1.20.0
k8s.gcr.io/pause:3.2
k8s.gcr.io/etcd:3.4.13-0
k8s.gcr.io/coredns:1.7.0
`, "\n"), "\n")},
		{"v1.21.0", strings.Split(strings.Trim(`
k8s.gcr.io/kube-apiserver:v1.21.0
k8s.gcr.io/kube-controller-manager:v1.21.0
k8s.gcr.io/kube-scheduler:v1.21.0
k8s.gcr.io/kube-proxy:v1.21.0
k8s.gcr.io/pause:3.4.1
k8s.gcr.io/etcd:3.4.13-0
k8s.gcr.io/coredns/coredns:v1.8.0
`, "\n"), "\n")},
	}
	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			v, err := semver.Make(strings.TrimPrefix(tc.version, "v"))
			if err != nil {
				t.Fatal(err)
			}
			want := tc.images
			got := essentials("k8s.gcr.io", v)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("images mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAuxiliary(t *testing.T) {
	want := []string{
		"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		"docker.io/kubernetesui/dashboard:v2.1.0",
		"docker.io/kubernetesui/metrics-scraper:v1.0.4",
	}
	got := auxiliary("")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("images mismatch (-want +got):\n%s", diff)
	}
}

func TestAuxiliaryMirror(t *testing.T) {
	want := []string{
		"test.mirror/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
		"test.mirror/kubernetesui/dashboard:v2.1.0",
		"test.mirror/kubernetesui/metrics-scraper:v1.0.4",
	}
	got := auxiliary("test.mirror")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("images mismatch (-want +got):\n%s", diff)
	}
}

func TestCNI(t *testing.T) {
	// images used by k8s.io/minikube/pkg/minikube/cni
	var testCases = []struct {
		name     string
		function func(string) string
	}{
		{"kindnet", KindNet},
		{"calico-deployment", CalicoDeployment},
		{"calico-daemonset", CalicoDaemonSet},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img := tc.function("")
			if img == "" {
				t.Errorf("no image")
			}
		})
	}
}
