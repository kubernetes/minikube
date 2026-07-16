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

import (
	"os"
	"strings"
	"testing"
)

func TestUpdateTimeouts(t *testing.T) {
	input, err := os.ReadFile("testdata/workflow_input.yml")
	if err != nil {
		t.Fatalf("ReadFile input: %v", err)
	}
	expected, err := os.ReadFile("testdata/workflow_expected.yml")
	if err != nil {
		t.Fatalf("ReadFile expected: %v", err)
	}

	actual, err := UpdateTimeouts(input, StepTimeouts{
		"Checkout":                2,
		"Test":                    15,
		"Run actions/setup-go@v6": 3,
		"Run make build":          8,
		"Docs":                    4,
		"Already tuned":           3,
	})
	if err != nil {
		t.Fatalf("UpdateTimeouts: %v", err)
	}

	if string(actual) != string(expected) {
		t.Fatalf("updated workflow mismatch (-want +got):\n%s", unifiedDiff(string(expected), string(actual)))
	}
}

func TestUpdateTimeoutsErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "invalid yaml",
			input:   "jobs: [",
			wantErr: "did not find expected node content",
		},
		{
			name:    "missing jobs",
			input:   "name: CI\n",
			wantErr: `get "jobs": not found`,
		},
		{
			name:    "jobs has wrong type",
			input:   "jobs: []\n",
			wantErr: "jobs: got sequence, want mapping",
		},
		{
			name: "steps has wrong type",
			input: `name: CI
jobs:
  build:
    steps: {}
`,
			wantErr: `job "build" steps: got mapping, want sequence`,
		},
		{
			name: "step name has wrong type",
			input: `name: CI
jobs:
  build:
    steps:
      - name:
          nested: value
        run: make test
`,
			wantErr: `job "build" step: get "name": got mapping, want scalar`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := UpdateTimeouts([]byte(tt.input), StepTimeouts{}); err == nil {
				t.Fatalf("UpdateTimeouts succeeded, want error containing %q", tt.wantErr)
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("UpdateTimeouts error = %q, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func unifiedDiff(want, got string) string {
	wantLines := strings.SplitAfter(want, "\n")
	gotLines := strings.SplitAfter(got, "\n")
	var b strings.Builder
	b.WriteString("--- want\n")
	b.WriteString("+++ got\n")
	for _, line := range wantLines {
		b.WriteString("-")
		b.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			b.WriteString("\n")
		}
	}
	for _, line := range gotLines {
		b.WriteString("+")
		b.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}
