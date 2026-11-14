/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package ssh

import (
	"encoding/json"
	"testing"
)

// The driver is created with an empty config and loaded from the machine
// config (see LocalClient.NewHost), so the container runtime must be restored
// from the machine config.
func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		expectRuntime string
	}{
		{
			name:          "containerd",
			json:          `{"MachineName":"minikube","ContainerRuntime":"containerd"}`,
			expectRuntime: "containerd",
		},
		{
			name:          "crio",
			json:          `{"MachineName":"minikube","ContainerRuntime":"crio"}`,
			expectRuntime: "CRI-O",
		},
		{
			name:          "docker",
			json:          `{"MachineName":"minikube","ContainerRuntime":"docker"}`,
			expectRuntime: "Docker",
		},
		{
			name:          "empty",
			json:          `{"MachineName":"minikube"}`,
			expectRuntime: "Docker",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDriver(Config{})
			if err := json.Unmarshal([]byte(tc.json), d); err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}
			if d.runtime.Name() != tc.expectRuntime {
				t.Errorf("expected runtime name %q, but got %q", tc.expectRuntime, d.runtime.Name())
			}
		})
	}
}

func TestUnmarshalJSONInvalidRuntime(t *testing.T) {
	d := NewDriver(Config{})
	if err := json.Unmarshal([]byte(`{"ContainerRuntime":"invalid"}`), d); err == nil {
		t.Errorf("expected error for invalid container runtime")
	}
}
