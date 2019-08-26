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

package lock

import "testing"

func TestGetMutexName(t *testing.T) {
	var tests = []struct {
		description string
		path        string
		expected    string
	}{
		{
			description: "standard",
			path:        "/foo/bar",
			expected:    "foo-bar",
		},
		{
			description: "deep directory",
			path:        "/foo/bar/baz/bat",
			expected:    "baz-bat",
		},
		{
			description: "underscores",
			path:        "/foo_bar/baz",
			expected:    "foo-bar-baz",
		},
		{
			description: "starts with number",
			path:        "/foo/2bar/baz",
			expected:    "m2bar-baz",
		},
		{
			description: "starts with punctuation",
			path:        "/.foo/bar",
			expected:    "foo-bar",
		},
		{
			description: "long filename",
			path:        "/very-very-very-very-very-very-very-very-long/bar",
			expected:    "very-very-very-very-very-very-very-very",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			got := getMutexName(tc.path)
			if got != tc.expected {
				t.Errorf("Unexpected mutex name for path %s. got: %s, expected: %s", tc.path, got, tc.expected)
			}
		})
	}
}
