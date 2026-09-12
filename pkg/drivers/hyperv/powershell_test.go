/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package hyperv

import "testing"

func TestParseLines(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   []string
	}{
		{
			name:  "multi-line",
			input: "Running\r\nOff\r\n",
			want:  []string{"Running", "Off"},
		},
		{
			name:  "single line no trailing newline",
			input: "Running",
			want:  []string{"Running"},
		},
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
		{
			name:  "with blank lines",
			input: "Running\n\nOff\n",
			want:  []string{"Running", "", "Off"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseLines(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("parseLines(%q) = %v, want %v", tc.input, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("parseLines(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}
