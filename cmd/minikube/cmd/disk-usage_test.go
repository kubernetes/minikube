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

package cmd

import (
	"testing"
)

const (
	test1 int64 = 1234567890
	test2 int64 = 12
	test3 int64 = 0
)

func TestInt64ToFloat32FormattedString(t *testing.T) {
	var tests = []struct {
		name                 string
		in                   int64
		expectedNumberOutput string
		expectedMapKeyOutput int
	}{
		{
			"long number",
			test1,
			"1.15",
			3,
		},
		{
			"short number",
			test2,
			"12.00",
			0,
		},
		{
			"zero",
			test3,
			"0.00",
			0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outNumber, outMapKey := int64ToFloat32FormattedString(tc.in)
			if outNumber != tc.expectedNumberOutput {
				t.Errorf("Error parsing float to string rounded to 2 digits\nExpected: %s\nGot: %s", tc.expectedNumberOutput, outNumber)
			}
			if outMapKey != tc.expectedMapKeyOutput {
				t.Errorf("Error getting decimal prefix map key\nExpected: %d\nGot: %d", tc.expectedMapKeyOutput, outMapKey)
			}
		})
	}
}
