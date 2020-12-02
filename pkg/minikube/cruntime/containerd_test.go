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
)

func TestAddRepoTagToImageName(t *testing.T) {
	var tests = []struct {
		imgName string
		want    string
	}{
		{"kubernetesui/dashboard:v2.0.3", "docker.io/kubernetesui/dashboard:v2.0.3"},
		{"kubernetesui/metrics-scraper:v1.0.4", "docker.io/kubernetesui/metrics-scraper:v1.0.4"},
		{"gcr.io/k8s-minikube/storage-provisioner:v4", "gcr.io/k8s-minikube/storage-provisioner:v4"},
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
