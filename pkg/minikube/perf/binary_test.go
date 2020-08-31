/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package perf

import (
	"strings"
	"testing"
)

func TestBinaryName(t *testing.T) {
	tests := []struct {
		expected string
		binary   Binary
	}{
		{
			expected: "foo",
			binary:   Binary{path: "foo", pr: 0},
		},
		{
			expected: "Minikube (PR 1)",
			binary:   Binary{path: "bar", pr: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			name := test.binary.Name()
			if name != test.expected {
				t.Fatalf("Binary name(%v) doesn't match expected name(%v)", name, test.expected)
			}
		})

	}
}

func TestNewBinary(t *testing.T) {
	tests := []struct {
		input, prNum string
		errExpected  bool
	}{
		{
			input: "42",
			prNum: "42",
		},
		{
			input:       prPrefix + "XYZ",
			errExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			bin, err := NewBinary(test.input)
			if err == nil && test.errExpected {
				t.Fatalf("Input %v returned unexpected error", test.input)
			}
			if test.errExpected {
				return
			}
			if bin == nil {
				t.Fatalf("Input string(%v) returned unexpected empty binary", test.input)
			}
			if !strings.Contains(bin.path, test.prNum) {
				t.Fatalf("Binary path(%v) doesn't contain expected(%v)", bin.path, test.prNum)
			}
		})
	}
}
