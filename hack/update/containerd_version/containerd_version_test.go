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

func TestContainerdNerdctlCompatible(t *testing.T) {
	tests := []struct {
		name       string
		containerd string
		nerdctl    string
		want       bool
	}{
		{name: "same major.minor", containerd: "2.2.1", nerdctl: "2.2.1", want: true},
		{name: "same major.minor, containerd patch ahead", containerd: "2.2.3", nerdctl: "2.2.0", want: true},
		{name: "same major.minor, nerdctl patch ahead", containerd: "2.2.0", nerdctl: "2.2.1", want: true},
		{name: "containerd major.minor behind nerdctl", containerd: "2.1.4", nerdctl: "2.2.1", want: true},
		{name: "containerd major.minor ahead of nerdctl", containerd: "2.3.0", nerdctl: "2.2.1", want: false},
		{name: "containerd a major ahead", containerd: "3.0.0", nerdctl: "2.2.1", want: false},
		{name: "old nerdctl, newer containerd minor", containerd: "1.7.26", nerdctl: "1.7.25", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containerdNerdctlCompatible(tt.containerd, tt.nerdctl); got != tt.want {
				t.Errorf("containerdNerdctlCompatible(%q, %q) = %v, want %v", tt.containerd, tt.nerdctl, got, tt.want)
			}
		})
	}
}

func TestMajorMinor(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{version: "2.2.1", want: "2.2"},
		{version: "1.7.25", want: "1.7"},
		{version: "2.0", want: "2.0"},
		{version: "3", want: "3"},
	}
	for _, tt := range tests {
		if got := majorMinor(tt.version); got != tt.want {
			t.Errorf("majorMinor(%q) = %q, want %q", tt.version, got, tt.want)
		}
	}
}

func TestVersionAhead(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{a: "2.2", b: "2.1", want: true},
		{a: "2.1", b: "2.2", want: false},
		{a: "2.2", b: "2.2", want: false},
		{a: "3.0", b: "2.9", want: true},
		{a: "2.10", b: "2.9", want: true}, // numeric, not lexicographic
		{a: "2.9", b: "2.10", want: false},
	}
	for _, tt := range tests {
		if got := versionAhead(tt.a, tt.b); got != tt.want {
			t.Errorf("versionAhead(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
