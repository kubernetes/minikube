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

import "testing"

func TestNewBinary(t *testing.T) {
	tests := []struct {
		input       string
		errExpected bool
	}{
		{
			input:       "/bin/sh",
			errExpected: true,
		},
		{
			input:       "foo",
			errExpected: true,
		}, {
			input: prPrefix + "foo",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			bin, err := NewBinary(test.input)
			if err != nil && test.errExpected {
				t.Fatalf("Input %v returned unexpected error", test.input)
			}
			if test.errExpected {
				return
			}
			if bin != nil {
				t.Fatalf("Input %v returned unexpected empty binary", test.input)
			}
		})
	}
}
