/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

func TestParsePath(t *testing.T) {
	var passedCases = []struct {
		path         string
		expectedNode string
		expectedPath string
	}{

		{"", "", ""},
		{":", "", ":"},
		{":/a", "", ":/a"},
		{":a", "", ":a"},
		{"minikube:", "", "minikube:"},
		{"minikube:./a", "", "minikube:./a"},
		{"minikube:a", "", "minikube:a"},
		{"minikube::a", "", "minikube::a"},
		{"./a", "", "./a"},
		{"./a/b", "", "./a/b"},
		{"a", "", "a"},
		{"a/b", "", "a/b"},
		{"/a", "", "/a"},
		{"/a/b", "", "/a/b"},
		{"./:a/b", "", "./:a/b"},
		{"c:\\a", "", "c:\\a"},
		{"c:\\a\\b", "", "c:\\a\\b"},
		{"minikube:/a", "minikube", "/a"},
		{"minikube:/a/b", "minikube", "/a/b"},
		{"minikube:/a/b:c", "minikube", "/a/b:c"},
	}

	for _, c := range passedCases {
		rp := newRemotePath(c.path)
		expected := remotePath{
			node: c.expectedNode,
			path: c.expectedPath,
		}
		if *rp != expected {
			t.Errorf("parsePath \"%s\" expected: %q, got: %q", c.path, expected, *rp)
		}
	}
}

func TestSetDstFileNameFromSrc(t *testing.T) {
	cases := []struct {
		src  string
		dst  string
		want string
	}{
		{"./a/b", "/c/", "/c/b"},
		{"./a/b", "node:/c/", "node:/c/b"},
		{"./a", "/c/", "/c/a"},
		{"", "/c/", "/c/"},
		{"./a/b", "", ""},
		{"./a/b", "/c", "/c"},
		{"./a/", "/c/", "/c/"},
	}

	for _, c := range cases {
		got := setDstFileNameFromSrc(c.dst, c.src)
		if c.want != got {
			t.Fatalf("wrong dst path for src=%s & dst=%s. want: %q, got: %q", c.src, c.dst, c.want, got)
		}
	}
}
