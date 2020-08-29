// +build darwin

/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package hyperkit

import (
	"testing"
)

func Test_portExtraction(t *testing.T) {
	tests := []struct {
		name    string
		ports   []string
		want    []int
		wantErr error
	}{
		{
			"valid_empty",
			[]string{},
			[]int{},
			nil,
		},
		{
			"valid_list",
			[]string{"10", "20", "30"},
			[]int{10, 20, 30},
			nil,
		},
		{
			"invalid",
			[]string{"8080", "not_an_integer"},
			nil,
			InvalidPortNumberError("not_an_integer"),
		},
	}

	for _, tt := range tests {
		d := NewDriver("", "")
		d.VSockPorts = tt.ports
		got, gotErr := d.extractVSockPorts()
		if !testEq(got, tt.want) {
			t.Errorf("extractVSockPorts() got: %v, want: %v", got, tt.want)
		}
		if gotErr != tt.wantErr {
			t.Errorf("extractVSockPorts() gotErr: %s, wantErr: %s", gotErr.Error(), tt.wantErr.Error())
		}
	}
}

func testEq(a, b []int) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
