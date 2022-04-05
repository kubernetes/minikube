/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package cruntime

import (
	"testing"

	"k8s.io/minikube/pkg/version"
)

func TestAddRepoTagToImageName(t *testing.T) {
	var tests = []struct {
		imgName string
		want    string
	}{
		{"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(), "gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion()},
	}
	for _, tc := range tests {
		t.Run(tc.imgName, func(t *testing.T) {
			got := addRepoTagToImageName(tc.imgName)
			if got != tc.want {
				t.Errorf("expected image name to be: %q but got %q", tc.want, got)
			}
		})
	}
}

func TestParseContainerdVersion(t *testing.T) {
	var tests = []struct {
		version string
		want    string
	}{
		{"containerd github.com/containerd/containerd v1.2.0 c4446665cb9c30056f4998ed953e6d4ff22c7c39", "1.2.0"},
		{"containerd github.com/containerd/containerd v1.2.1-rc.0 de1f167ab96338a9f5c2b17347abf84bdf1dd411", "1.2.1-rc.0"},
		{"containerd github.com/containerd/containerd 1.4.4-0ubuntu1 ", "1.4.4-0ubuntu1"},
		{"containerd github.com/containerd/containerd 1.5.2-0ubuntu1~21.04.2 ", "1.5.2-0ubuntu1"},
		{"containerd github.com/containerd/containerd 1.5.4~ds1 1.5.4~ds1-1", "1.5.4"},
	}
	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			got, err := parseContainerdVersion(tc.version)
			if err != nil {
				t.Fatalf("parse(%s): %v", tc.version, err)
			}
			if got != tc.want {
				t.Errorf("expected version to be: %q but got %q", tc.want, got)
			}
		})
	}
}
