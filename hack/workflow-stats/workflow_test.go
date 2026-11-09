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

	"github.com/google/go-cmp/cmp"
)

func TestUpdateTimeouts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		timeouts StepTimeouts
	}{
		{
			name:     "add timeout",
			input:    "testdata/add_timeout_input.yml",
			expected: "testdata/add_timeout_expected.yml",
			timeouts: StepTimeouts{
				"Test": 15,
			},
		},
		{
			name:     "update timeout",
			input:    "testdata/update_timeout_input.yml",
			expected: "testdata/update_timeout_expected.yml",
			timeouts: StepTimeouts{
				"Test": 15,
			},
		},
		{
			name:     "unnamed steps",
			input:    "testdata/unnamed_steps_input.yml",
			expected: "testdata/unnamed_steps_expected.yml",
			timeouts: StepTimeouts{
				"Run actions/setup-go@v6": 3,
				"Run make build":          8,
			},
		},
		{
			name:     "timeout comments",
			input:    "testdata/timeout_comments_input.yml",
			expected: "testdata/timeout_comments_expected.yml",
			timeouts: StepTimeouts{
				"Test": 15,
			},
		},
		{
			name:     "add timeout keeps comments",
			input:    "testdata/add_timeout_comments_input.yml",
			expected: "testdata/add_timeout_comments_expected.yml",
			timeouts: StepTimeouts{
				"Test": 15,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := os.ReadFile(tt.input)
			if err != nil {
				t.Fatalf("ReadFile input: %v", err)
			}
			expected, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatalf("ReadFile expected: %v", err)
			}

			actual, err := UpdateTimeouts(input, tt.timeouts)
			if err != nil {
				t.Fatalf("UpdateTimeouts: %v", err)
			}

			if diff := cmp.Diff(string(expected), string(actual)); diff != "" {
				t.Fatalf("updated workflow mismatch (-want +got):\n%s", diff)
			}
		})
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
