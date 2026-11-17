/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package util

import "testing"

func TestHasVolumeNameWindows(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{`C:\a\b`, true},
		{`C:/a/b`, true},
		{`c:/a`, true},
		{`\\server\share\a`, true},
		{`/a/b`, false},
		{`a/b`, false},
		{`minikube:/a`, false},
	}
	for _, c := range cases {
		if got := HasVolumeName(c.path); got != c.want {
			t.Errorf("HasVolumeName(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
