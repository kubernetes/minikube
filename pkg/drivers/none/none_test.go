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

package none

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		expectRuntime   string
		expectNoRuntime bool
	}{
		{
			name:          "containerd",
			json:          `{"MachineName":"minikube","ContainerRuntime":"containerd"}`,
			expectRuntime: "containerd",
		},
		{
			name:          "docker",
			json:          `{"MachineName":"minikube","ContainerRuntime":"docker"}`,
			expectRuntime: "Docker",
		},
		{
			name:            "empty",
			json:            `{"MachineName":"minikube"}`,
			expectNoRuntime: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var d Driver
			if err := json.Unmarshal([]byte(tc.json), &d); err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}

			if tc.expectNoRuntime {
				if d.runtime != nil {
					t.Errorf("expected runtime to be nil, but got %v", d.runtime)
				}
				return
			}

			if d.runtime == nil {
				t.Fatalf("expected runtime to be initialized, but got nil")
			}

			if d.runtime.Name() != tc.expectRuntime {
				t.Errorf("expected runtime name %q, but got %q", tc.expectRuntime, d.runtime.Name())
			}
		})
	}
}
