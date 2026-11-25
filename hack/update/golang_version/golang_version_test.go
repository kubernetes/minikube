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

package main

import "testing"

func TestParseGoBuildImageVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantGo    string
		wantK8s   string
		wantError bool
	}{
		{name: "valid", input: "v1.35.0-go1.27.1-bookworm.0\n", wantGo: "1.27.1", wantK8s: "v1.35.0"},
		{name: "missing separator", input: "v1.35.0", wantError: true},
		{name: "missing Go version", input: "v1.35.0-go-", wantError: true},
		{name: "unexpected version", input: "v1.35.0-rust1.98", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goVersion, k8sVersion, err := parseGoBuildImageVersion(tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && (goVersion != tt.wantGo || k8sVersion != tt.wantK8s) {
				t.Fatalf("got Go %q and Kubernetes %q, want %q and %q", goVersion, k8sVersion, tt.wantGo, tt.wantK8s)
			}
		})
	}
}
