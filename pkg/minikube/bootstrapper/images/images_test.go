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
registry.k8s.io/kube-apiserver:v1.18.0
registry.k8s.io/kube-controller-manager:v1.18.0
registry.k8s.io/kube-scheduler:v1.18.0
registry.k8s.io/kube-proxy:v1.18.0
registry.k8s.io/pause:3.2
registry.k8s.io/etcd:3.4.3-0
registry.k8s.io/coredns:1.6.7
`, "\n"), "\n")},
		{"v1.19.0", strings.Split(strings.Trim(`
registry.k8s.io/kube-apiserver:v1.19.0
registry.k8s.io/kube-controller-manager:v1.19.0
registry.k8s.io/kube-scheduler:v1.19.0
registry.k8s.io/kube-proxy:v1.19.0
registry.k8s.io/pause:3.2
registry.k8s.io/etcd:3.4.9-1
registry.k8s.io/coredns:1.7.0
`, "\n"), "\n")},
		{"v1.20.0", strings.Split(strings.Trim(`
registry.k8s.io/kube-apiserver:v1.20.0
registry.k8s.io/kube-controller-manager:v1.20.0
registry.k8s.io/kube-scheduler:v1.20.0
registry.k8s.io/kube-proxy:v1.20.0
registry.k8s.io/pause:3.2
registry.k8s.io/etcd:3.4.13-0
registry.k8s.io/coredns:1.7.0
`, "\n"), "\n")},
		{"v1.21.0", strings.Split(strings.Trim(`
registry.k8s.io/kube-apiserver:v1.21.0
registry.k8s.io/kube-controller-manager:v1.21.0
registry.k8s.io/kube-scheduler:v1.21.0
registry.k8s.io/kube-proxy:v1.21.0
registry.k8s.io/pause:3.4.1
registry.k8s.io/etcd:3.4.13-0
registry.k8s.io/coredns/coredns:v1.8.0
`, "\n"), "\n")},
		{"v1.22.0", strings.Split(strings.Trim(`
registry.k8s.io/kube-apiserver:v1.22.0
registry.k8s.io/kube-controller-manager:v1.22.0
registry.k8s.io/kube-scheduler:v1.22.0
registry.k8s.io/kube-proxy:v1.22.0
registry.k8s.io/pause:3.5
registry.k8s.io/etcd:3.5.0-0
registry.k8s.io/coredns/coredns:v1.8.4
`, "\n"), "\n")},
	}
	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			v, err := semver.Make(strings.TrimPrefix(tc.version, "v"))
			if err != nil {
				t.Fatal(err)
			}
			want := tc.images
			got := essentials("registry.k8s.io", v)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("images mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEssentialsAliyunMirror(t *testing.T) {
	var testCases = []struct {
		version string
		images  []string
	}{

		{"v1.21.0", strings.Split(strings.Trim(`
registry.cn-hangzhou.aliyuncs.com/google_containers/kube-apiserver:v1.21.0
registry.cn-hangzhou.aliyuncs.com/google_containers/kube-controller-manager:v1.21.0
registry.cn-hangzhou.aliyuncs.com/google_containers/kube-scheduler:v1.21.0
registry.cn-hangzhou.aliyuncs.com/google_containers/kube-proxy:v1.21.0
registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.4.1
registry.cn-hangzhou.aliyuncs.com/google_containers/etcd:3.4.13-0
registry.cn-hangzhou.aliyuncs.com/google_containers/coredns:v1.8.0
`, "\n"), "\n")},
		{"v1.22.0", strings.Split(strings.Trim(`
registry.cn-hangzhou.aliyuncs.com/google_containers/kube-apiserver:v1.22.0
registry.cn-hangzhou.aliyuncs.com/google_containers/kube-controller-manager:v1.22.0
registry.cn-hangzhou.aliyuncs.com/google_containers/kube-scheduler:v1.22.0
registry.cn-hangzhou.aliyuncs.com/google_containers/kube-proxy:v1.22.0
registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.5
registry.cn-hangzhou.aliyuncs.com/google_containers/etcd:3.5.0-0
registry.cn-hangzhou.aliyuncs.com/google_containers/coredns:v1.8.4
`, "\n"), "\n")},
	}
	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			v, err := semver.Make(strings.TrimPrefix(tc.version, "v"))
			if err != nil {
				t.Fatal(err)
			}
			want := tc.images
			got := essentials("registry.cn-hangzhou.aliyuncs.com/google_containers", v)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("images mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAuxiliary(t *testing.T) {
	want := []string{
		"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
	}
	got := auxiliary("")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("images mismatch (-want +got):\n%s", diff)
	}
}

func TestAuxiliaryMirror(t *testing.T) {
	want := []string{
		"test.mirror/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(),
	}
	got := auxiliary("test.mirror")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("images mismatch (-want +got):\n%s", diff)
	}
}

func TestAuxiliaryAliyunMirror(t *testing.T) {
	want := []string{
		"registry.cn-hangzhou.aliyuncs.com/google_containers/storage-provisioner:" + version.GetStorageProvisionerVersion(),
	}
	got := auxiliary("registry.cn-hangzhou.aliyuncs.com/google_containers")
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

func TestTagFromLastMinor(t *testing.T) {
	tests := []struct {
		verString   string
		imageName   string
		expectedTag string
	}{
		{"1.16.50", "coredns", "1.6.2"},
		{"1.16.50", "etcd", "3.3.15-0"},
		{"1.16.50", "pause", "3.1"},
		{"1.16.50", "fake", "default"},
	}

	for _, tc := range tests {
		v, err := semver.Parse(tc.verString)
		if err != nil {
			t.Errorf("failed to parse version to semver: %v", err)
			continue
		}
		got := tagFromLastMinor(v, tc.imageName, "default")
		if tc.expectedTag != got {
			t.Errorf("tagFromLastMinor(%s, %s, default) = %s; want = %s", tc.verString, tc.imageName, got, tc.expectedTag)
		}
	}
}
