/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package cni

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

func TestChooseDefault(t *testing.T) {
	tests := []struct {
		driver           string
		containerRuntime string
		version          string
		want             string
	}{
		{"docker", "docker", "v1.23.0", "Disabled"},
		{"docker", "docker", "v1.27.0", "bridge CNI"},
		{"docker", "containerd", "v1.23.0", "CNI"},
		{"docker", "containerd", "v1.27.0", "CNI"},
		{"qemu", "docker", "v1.23.0", "Disabled"},
		{"qemu", "docker", "v1.27.0", "bridge CNI"},
		{"qemu", "containerd", "v1.23.0", "bridge CNI"},
		{"qemu", "containerd", "v1.27.0", "bridge CNI"},
	}
	for _, tc := range tests {
		cc := config.ClusterConfig{
			Driver: tc.driver,
			KubernetesConfig: config.KubernetesConfig{
				ContainerRuntime:  tc.containerRuntime,
				KubernetesVersion: tc.version,
			},
		}
		got := chooseDefault(cc).String()
		if got != tc.want {
			t.Errorf("chooseDefault(%+v) = %s; want = %s", cc, got, tc.want)
		}
	}
}
