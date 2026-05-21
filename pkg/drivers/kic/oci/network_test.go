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

package oci

import (
	"strings"
	"testing"
)

func TestDockerContainerIP(t *testing.T) {
	tests := []struct {
		name        string
		inspectOut  []string // lines returned by the mocked inspect call
		wantIPv4    string
		wantIPv6    string
		wantErr     bool
	}{
		{
			name:       "single network with ipv4",
			inspectOut: []string{"10.0.0.2,"},
			wantIPv4:   "10.0.0.2",
			wantIPv6:   "",
		},
		{
			name:       "single network with ipv4 and ipv6",
			inspectOut: []string{"10.0.0.2,fe80::1"},
			wantIPv4:   "10.0.0.2",
			wantIPv6:   "fe80::1",
		},
		{
			name:       "multiple networks returns first with ipv4",
			inspectOut: []string{"10.0.0.2,", "192.168.1.5,"},
			wantIPv4:   "10.0.0.2",
			wantIPv6:   "",
		},
		{
			name:       "first network has no ipv4 falls through to next",
			inspectOut: []string{",fe80::1", "10.0.0.2,"},
			wantIPv4:   "10.0.0.2",
			wantIPv6:   "",
		},
		{
			name:       "blank lines are skipped",
			inspectOut: []string{"", "  ", "10.0.0.2,"},
			wantIPv4:   "10.0.0.2",
			wantIPv6:   "",
		},
		{
			name:       "no network has ipv4",
			inspectOut: []string{",fe80::1"},
			wantErr:    true,
		},
		{
			name:       "empty output",
			inspectOut: []string{},
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := dockerContainerInspect
			t.Cleanup(func() { dockerContainerInspect = orig })

			dockerContainerInspect = func(_, _, _ string) ([]string, error) {
				return tc.inspectOut, nil
			}

			ipv4, ipv6, err := dockerContainerIP("docker", "minikube")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if err != nil {
				return
			}
			if ipv4 != tc.wantIPv4 {
				t.Errorf("ipv4: got %q, want %q", ipv4, tc.wantIPv4)
			}
			if ipv6 != tc.wantIPv6 {
				t.Errorf("ipv6: got %q, want %q", ipv6, tc.wantIPv6)
			}
		})
	}
}

func TestDockerContainerIPMalformedLine(t *testing.T) {
	orig := dockerContainerInspect
	t.Cleanup(func() { dockerContainerInspect = orig })

	// A line with more or fewer than 2 comma-separated fields should error.
	dockerContainerInspect = func(_, _, _ string) ([]string, error) {
		return []string{"10.0.0.2"}, nil // missing the comma separator
	}

	_, _, err := dockerContainerIP("docker", "minikube")
	if err == nil {
		t.Error("expected error for malformed inspect line, got nil")
	}
	if !strings.Contains(err.Error(), "container addresses should have 2 values") {
		t.Errorf("unexpected error message: %v", err)
	}
}
