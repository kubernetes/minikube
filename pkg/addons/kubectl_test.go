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

package addons

import (
	"context"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

func TestKubectlCommand(t *testing.T) {
	tests := []struct {
		description string
		files       []string
		enable      bool
		force       bool
		expected    string
	}{
		{
			description: "enable an addon",
			files:       []string{"a", "b"},
			enable:      true,
			expected:    "sudo KUBECONFIG=/var/lib/minikube/kubeconfig /var/lib/minikube/binaries/v1.17.0/kubectl apply -f a -f b",
		},
		{
			description: "disable an addon",
			files:       []string{"a", "b"},
			enable:      false,
			expected:    "sudo KUBECONFIG=/var/lib/minikube/kubeconfig /var/lib/minikube/binaries/v1.17.0/kubectl delete --ignore-not-found -f a -f b",
		},
		{
			description: "enable an addon",
			files:       []string{"a", "b"},
			enable:      true,
			force:       true,
			expected:    "sudo KUBECONFIG=/var/lib/minikube/kubeconfig /var/lib/minikube/binaries/v1.17.0/kubectl apply --force -f a -f b",
		},
		{
			description: "disable an addon",
			files:       []string{"a", "b"},
			enable:      false,
			force:       true,
			expected:    "sudo KUBECONFIG=/var/lib/minikube/kubeconfig /var/lib/minikube/binaries/v1.17.0/kubectl delete --force --ignore-not-found -f a -f b",
		},
	}

	cc := &config.ClusterConfig{
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion: "v1.17.0",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			command := kubectlCommand(context.Background(), cc, test.files, test.enable, test.force)
			actual := strings.Join(command.Args, " ")

			if actual != test.expected {
				t.Fatalf("expected does not match actual\nExpected: %s\nActual: %s", test.expected, actual)
			}
		})
	}
}
