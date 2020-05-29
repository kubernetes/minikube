/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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
	"testing"
)

func TestAverage(t *testing.T) {
	tests := []struct {
		description string
		nums        []float64
		expected    float64
	}{
		{
			description: "one number",
			nums:        []float64{4},
			expected:    4,
		}, {
			description: "multiple numbers",
			nums:        []float64{1, 4},
			expected:    2.5,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := average(test.nums)
			if actual != test.expected {
				t.Fatalf("actual output does not match expected output\nActual: %v\nExpected: %v", actual, test.expected)
			}
		})
	}
}
