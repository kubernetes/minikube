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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAuxiliary(t *testing.T) {
	want := []string{
		"k8s.gcr.io/kube-addon-manager:v9.0.2",
		"gcr.io/k8s-minikube/storage-provisioner:v1.8.1",
		"kubernetesui/dashboard:v2.0.0-beta8",
		"kubernetesui/metrics-scraper:v1.0.2",
	}
	got := Auxiliary("")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("images mismatch (-want +got):\n%s", diff)
	}
}

func TestAuxiliaryMirror(t *testing.T) {
	want := []string{
		"test.mirror/kube-addon-manager:v9.0.2",
		"test.mirror/storage-provisioner:v1.8.1",
		"test.mirror/dashboard:v2.0.0-beta8",
		"test.mirror/metrics-scraper:v1.0.2",
	}
	got := Auxiliary("test.mirror")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("images mismatch (-want +got):\n%s", diff)
	}
}
