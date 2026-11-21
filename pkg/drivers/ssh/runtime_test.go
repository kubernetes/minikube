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
	"strings"
	"testing"
)

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		expectRuntime string
		wantErr       bool
	}{
		{name: "containerd", raw: `{"MachineName":"minikube","ContainerRuntime":"containerd"}`, expectRuntime: "containerd"},
		{name: "crio", raw: `{"MachineName":"minikube","ContainerRuntime":"crio"}`, expectRuntime: "CRI-O"},
		{name: "docker", raw: `{"MachineName":"minikube","ContainerRuntime":"docker"}`, expectRuntime: "Docker"},
		{name: "absent is docker", raw: `{"MachineName":"minikube"}`, expectRuntime: "Docker"},
		{name: "invalid", raw: `{"ContainerRuntime":"invalid"}`, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDriver(Config{})
			err := json.Unmarshal([]byte(tc.raw), d)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error for invalid container runtime")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}
			if got := d.RuntimeName(); got != tc.expectRuntime {
				t.Errorf("runtime name = %q, want %q", got, tc.expectRuntime)
			}
		})
	}
}

func TestUseContainerRuntime(t *testing.T) {
	tests := []struct {
		name          string
		runtime       string
		expectRuntime string
		wantErr       bool
	}{
		{name: "containerd", runtime: "containerd", expectRuntime: "containerd"},
		{name: "crio", runtime: "crio", expectRuntime: "CRI-O"},
		{name: "cri-o", runtime: "cri-o", expectRuntime: "CRI-O"},
		{name: "docker", runtime: "docker", expectRuntime: "Docker"},
		{name: "empty keeps docker", runtime: "", expectRuntime: "Docker"},
		{name: "invalid", runtime: "invalid", wantErr: true, expectRuntime: "Docker"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDriver(Config{})
			err := d.UseContainerRuntime(tc.runtime)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error for invalid container runtime")
				}
			} else if err != nil {
				t.Fatalf("UseContainerRuntime(%q) failed: %v", tc.runtime, err)
			}
			if got := d.RuntimeName(); got != tc.expectRuntime {
				t.Errorf("runtime name = %q, want %q", got, tc.expectRuntime)
			}
			if tc.wantErr {
				if d.ContainerRuntime != "" {
					t.Errorf("ContainerRuntime = %q after error, want empty", d.ContainerRuntime)
				}
				return
			}
			if d.ContainerRuntime != tc.runtime {
				t.Errorf("ContainerRuntime = %q, want %q", d.ContainerRuntime, tc.runtime)
			}
		})
	}
}

func TestUseContainerRuntimePersistsName(t *testing.T) {
	d := NewDriver(Config{})
	if err := d.UseContainerRuntime("containerd"); err != nil {
		t.Fatal(err)
	}
	buf, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buf), `"ContainerRuntime":"containerd"`) {
		t.Errorf("marshaled driver = %s, want ContainerRuntime containerd", buf)
	}
}
