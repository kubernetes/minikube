/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package detect

import "testing"

func TestParseDockerHubRemaining(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		want       int
		shouldFail bool
	}{
		{
			name:   "ExampleFromDocs",
			header: "100;w=21600",
			want:   100,
		},
		{
			name:   "CommaSeparated",
			header: "5000;w=60,burst=5000",
			want:   5000,
		},
		{
			name:   "Whitespace",
			header: " 42 ; w=60 ",
			want:   42,
		},
		{
			name:       "NoNumericValue",
			header:     "w=21600",
			shouldFail: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDockerHubRemaining(tc.header)
			if tc.shouldFail {
				if err == nil {
					t.Fatalf("parseDockerHubRemaining(%q) expected error, got value %d", tc.header, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDockerHubRemaining(%q) unexpected error: %v", tc.header, err)
			}
			if got != tc.want {
				t.Fatalf("parseDockerHubRemaining(%q) = %d, want %d", tc.header, got, tc.want)
			}
		})
	}
}
