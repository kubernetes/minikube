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

package style

import "testing"

func TestLowPrefix(t *testing.T) {
	tests := []struct {
		expected    string
		description string
		style       Options
	}{
		{
			expected:    LowBullet,
			description: "empty prefix",
		},
		{
			expected:    "bar",
			style:       Options{LowPrefix: "bar"},
			description: "lowPrefix",
		},
		{
			expected:    LowBullet,
			style:       Options{Prefix: "foo"},
			description: "prefix without spaces",
		},
		{
			expected:    LowIndent,
			style:       Options{Prefix: "  foo"},
			description: "prefix with spaces",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := LowPrefix(test.style)
			if got != test.expected {
				t.Errorf("Expected %v but got %v", test.expected, got)
			}
		})
	}
}
