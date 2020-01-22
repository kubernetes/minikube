/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package ktmpl

import (
	"reflect"
	"testing"
)

func TestPrintMapInOrder(t *testing.T) {
	tests := []struct {
		description string
		m           map[string]string
		sep         string
		expected    []string
	}{
		{
			description: "single kv",
			sep:         ": ",
			m: map[string]string{
				"a": "1",
			},
			expected: []string{`a: "1"`},
		},
		{
			description: "two kv",
			sep:         "=",
			m: map[string]string{
				"b": "2",
				"a": "1",
			},
			expected: []string{`a="1"`, `b="2"`},
		},
		{
			description: "no kv",
			sep:         ",",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := printMapInOrder(test.m, test.sep)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Actual: %v, Expected: %v", actual, test.expected)
			}
		})
	}
}
