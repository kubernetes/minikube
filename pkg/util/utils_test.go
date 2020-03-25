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
	"testing"

	"github.com/blang/semver"
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

func TestCalculateSizeInMB(t *testing.T) {
	testData := []struct {
		size           string
		expectedNumber int
	}{
		{"1024kb", 1},
		{"1024KB", 1},
		{"1024mb", 1024},
		{"1024b", 0},
		{"1g", 1024},
	}

	for _, tt := range testData {
		number, err := CalculateSizeInMB(tt.size)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if number != tt.expectedNumber {
			t.Fatalf("Expected '%d' but got '%d' from size '%s'", tt.expectedNumber, number, tt.size)
		}
	}
}

func TestParseKubernetesVersion(t *testing.T) {
	version, err := ParseKubernetesVersion("v1.8.0-alpha.5")
	if err != nil {
		t.Fatalf("Error parsing version: %v", err)
	}
	if version.NE(semver.MustParse("1.8.0-alpha.5")) {
		t.Errorf("Expected: %s, Actual:%s", "1.8.0-alpha.5", version)
	}
}
