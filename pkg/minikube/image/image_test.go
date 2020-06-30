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

package image

import "testing"

func TestTag(t *testing.T) {
	tcs := []struct {
		image    string
		expected string
	}{
		{
			image:    "image:tag@sha256:digest",
			expected: "image:tag",
		}, {
			image:    "image:tag",
			expected: "image:tag",
		}, {
			image:    "image@sha256:digest",
			expected: "image@sha256:digest",
		}, {
			image:    "image",
			expected: "image",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.image, func(t *testing.T) {
			actual := Tag(tc.image)
			if actual != tc.expected {
				t.Errorf("actual does not match expected\nActual:%v\nExpected:%v\n", actual, tc.expected)
			}
		})
	}
}
