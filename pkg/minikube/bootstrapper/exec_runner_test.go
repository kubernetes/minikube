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

package bootstrapper

import (
	"reflect"
	"testing"
)

func TestSplitCommandString(t *testing.T) {
	testCases := []struct {
		input       string
		wantCommand string
		wantArgs    []string
	}{
		{"", "", []string{}},
		{"ls", "ls", []string{}},
		{"ls -l", "ls", []string{"-l"}},
		{"ls -hl", "ls", []string{"-hl"}},
		{"ls -h -l", "ls", []string{"-h", "-l"}},
		{"ls -hl foo/", "ls", []string{"-hl", "foo/"}},
		{"ls bar/", "ls", []string{"bar/"}},
	}

	for _, tc := range testCases {
		gotCommand, gotArgs := splitCommandString(tc.input)

		if gotCommand != tc.wantCommand {
			t.Errorf("For Test Case %v, Expected Command %v, Got Command %v", tc.input, tc.wantCommand, gotCommand)
		}

		if len(tc.wantArgs) != 0 && reflect.DeepEqual(gotArgs, tc.wantArgs) == false {
			t.Log(reflect.DeepEqual(gotArgs, tc.wantArgs))
			t.Errorf("For Test Case %v, Expected Args %v, Got Args %v", tc.input, tc.wantArgs, gotArgs)
		}
	}
}
