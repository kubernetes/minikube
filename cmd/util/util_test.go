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

package util

import (
	"os"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

func TestGetKubeConfigPath(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{
			input: "/home/fake/.kube/.kubeconfig",
			want:  "/home/fake/.kube/.kubeconfig",
		},
		{
			input: "/home/fake/.kube/.kubeconfig:/home/fake2/.kubeconfig",
			want:  "/home/fake/.kube/.kubeconfig",
		},
	}

	for _, test := range tests {
		os.Setenv(clientcmd.RecommendedConfigPathEnvVar, test.input)
		if result := GetKubeConfigPath(); result != test.want {
			t.Errorf("Expected first splitted chunk, got: %s", result)
		}
	}
}
