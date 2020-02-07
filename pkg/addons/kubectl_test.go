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
	"fmt"
	"path"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/vmpath"
)

func TestKubectlCommand(t *testing.T) {
	tests := []struct {
		description string
		files       []string
		enable      bool
		expected    string
	}{
		{
			description: "enable an addon",
			files:       []string{"a", "b"},
			enable:      true,
			expected:    fmt.Sprintf("sudo KUBECONFIG=%s %s apply -f a -f b", path.Join(vmpath.GuestPersistentDir, "kubeconfig"), path.Join(vmpath.GuestPersistentDir, "v1.17.0".kubectl)),
		}, {
			description: "disable an addon",
			files:       []string{"a", "b"},
			enable:      false,
			expected:    fmt.Sprintf("sudo KUBECONFIG=%s %s apply -f a -f b", path.Join(vmpath.GuestPersistentDir, "kubeconfig"), path.Join(vmpath.GuestPersistentDir, "v1.17.0".kubectl)),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			originalK8sVersion := k8sVersion
			defer func() { k8sVersion = originalK8sVersion }()
			k8sVersion = func(_ string) (string, error) {
				return "v1.17.0", nil
			}

			command, err := kubectlCommand("", test.files, test.enable)
			if err != nil {
				t.Fatalf("error getting kubectl command: %v", err)
			}
			actual := strings.Join(command.Args, " ")

			if actual != test.expected {
				t.Fatalf("expected does not match actual\nExpected: %s\nActual: %s", test.expected, actual)
			}
		})
	}
}
