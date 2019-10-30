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

package util

import (
	"strings"
	"testing"
)

func TestGetBinaryDownloadURL(t *testing.T) {
	testData := []struct {
		version     string
		platform    string
		expectedURL string
	}{
		{"v0.0.1", "linux", "https://storage.googleapis.com/minikube/releases/v0.0.1/minikube-linux-amd64"},
		{"v0.0.1", "darwin", "https://storage.googleapis.com/minikube/releases/v0.0.1/minikube-darwin-amd64"},
		{"v0.0.1", "windows", "https://storage.googleapis.com/minikube/releases/v0.0.1/minikube-windows-amd64.exe"},
	}

	for _, tt := range testData {
		url := GetBinaryDownloadURL(tt.version, tt.platform)
		if url != tt.expectedURL {
			t.Fatalf("Expected '%s' but got '%s'", tt.expectedURL, url)
		}
	}

}

func TestReplaceChars(t *testing.T) {
	testData := []struct {
		src         []string
		replacer    *strings.Replacer
		expectedRes []string
	}{
		{[]string{"abc%def", "%Y%"}, strings.NewReplacer("%", "X"), []string{"abcXdef", "XYX"}},
	}

	for _, tt := range testData {
		res := ReplaceChars(tt.src, tt.replacer)
		for i, val := range res {
			if val != tt.expectedRes[i] {
				t.Fatalf("Expected '%s' but got '%s'", tt.expectedRes, res)
			}
		}
	}
}

func TestConcatStrings(t *testing.T) {
	testData := []struct {
		src         []string
		prefix      string
		postfix     string
		expectedRes []string
	}{
		{[]string{"abc", ""}, "xx", "yy", []string{"xxabcyy", "xxyy"}},
		{[]string{"abc", ""}, "", "", []string{"abc", ""}},
	}

	for _, tt := range testData {
		res := ConcatStrings(tt.src, tt.prefix, tt.postfix)
		for i, val := range res {
			if val != tt.expectedRes[i] {
				t.Fatalf("Expected '%s' but got '%s'", tt.expectedRes, res)
			}
		}
	}
}
