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

package cmd

import "testing"

func TestParsePathWindows(t *testing.T) {
	cases := []struct {
		path         string
		expectedNode string
		expectedPath string
	}{
		{`C:/a/b`, "", `C:/a/b`},
		{`c:/a`, "", `c:/a`},
		{`C:\a\b`, "", `C:\a\b`},
		{"minikube:/a", "minikube", "/a"},
		{"minikube-m02:/a/b", "minikube-m02", "/a/b"},
	}
	for _, c := range cases {
		rp := newRemotePath(c.path)
		expected := remotePath{node: c.expectedNode, path: c.expectedPath}
		if *rp != expected {
			t.Errorf("parsePath %q expected: %q, got: %q", c.path, expected, *rp)
		}
	}
}

func TestSetDstFileNameFromSrcWindows(t *testing.T) {
	cases := []struct {
		src  string
		dst  string
		want string
	}{
		{"C:/a/b", "/c/", "/c/b"},
		{"C:/a/b", "node:/c/", "node:/c/b"},
	}
	for _, c := range cases {
		if got := setDstFileNameFromSrc(c.dst, c.src); got != c.want {
			t.Errorf("setDstFileNameFromSrc(%q, %q) = %q, want %q", c.dst, c.src, got, c.want)
		}
	}
}
