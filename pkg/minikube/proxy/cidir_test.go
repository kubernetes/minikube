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

package proxy

import (
	"testing"
)

var bTests = []struct {
	ip             string
	block          string
	expectedResult bool
	expectedErr    error
}{
	{"", "192.168.0.1/32", false, nil},
	{"192.168.0.1", "192.168.0.1/32", true, nil},
	{"192.168.0.2", "192.168.0.1/32", false, nil},
	{"192.168.0.1", "192.168.0.1/18", true, nil},
	{"abcd", "192.168.0.1/18", false, nil},
}

func TestIsInBlock(t *testing.T) {

	for _, tt := range bTests {
		actualR, actualErr := isInBlock(tt.ip, tt.block)
		if actualR != tt.expectedResult {
			t.Errorf("isInBlock(%s,%s): expected %t, actual %t", tt.ip, tt.block, tt.expectedResult, actualR)
		}
		if actualErr != tt.expectedErr {
			t.Errorf("isInBlock(%s,%s): expected error %s, err %s", tt.ip, tt.block, tt.expectedErr, actualErr)
		}
	}

}
