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
	"os"
	"os/user"
	"syscall"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
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
		url := GetBinaryDownloadURL(tt.version, tt.platform, "amd64")
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

func TestChownR(t *testing.T) {
	testDir := t.TempDir()
	if _, err := os.Create(testDir + "/TestChownR"); err != nil {
		return
	}

	cases := []struct {
		name          string
		uid           int
		gid           int
		expectedError bool
	}{
		{
			name:          "normal",
			uid:           os.Getuid(),
			gid:           os.Getgid(),
			expectedError: false,
		},
		{
			name:          "invalid uid",
			uid:           2147483647,
			gid:           os.Getgid(),
			expectedError: true,
		},
		{
			name:          "invalid gid",
			uid:           os.Getuid(),
			gid:           2147483647,
			expectedError: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ChownR(testDir+"/TestChownR", c.uid, c.gid)
			fileInfo, _ := os.Stat(testDir + "/TestChownR")
			fileSys := fileInfo.Sys()
			if (nil != err) != c.expectedError || ((false == c.expectedError) && (fileSys.(*syscall.Stat_t).Gid != uint32(c.gid) || fileSys.(*syscall.Stat_t).Uid != uint32(c.uid))) {
				t.Errorf("expectedError: %v, got: %v", c.expectedError, err)
			}
		})
	}
}

func TestMaybeChownDirRecursiveToMinikubeUser(t *testing.T) {
	testDir := t.TempDir()
	if _, err := os.Create(testDir + "/TestChownR"); nil != err {
		return
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		t.Setenv("CHANGE_MINIKUBE_NONE_USER", "1")
	}

	if os.Getenv("SUDO_USER") == "" {
		user, err := user.Current()
		if nil != err {
			t.Error("fail to get user")
		}
		t.Setenv("SUDO_USER", user.Username)
	}

	cases := []struct {
		name          string
		dir           string
		expectedError bool
	}{
		{
			name:          "normal",
			dir:           testDir,
			expectedError: false,
		},
		{
			name:          "invalid dir",
			dir:           "./utils_test",
			expectedError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := MaybeChownDirRecursiveToMinikubeUser(c.dir)
			if (nil != err) != c.expectedError {
				t.Errorf("expectedError: %v, got: %v", c.expectedError, err)
			}
		})
	}
}

func TestRemoveDuplicateStrings(t *testing.T) {
	testCases := []struct {
		desc  string
		slice []string
		want  []string
	}{
		{
			desc:  "NoDuplicates",
			slice: []string{"alpha", "bravo", "charlie"},
			want:  []string{"alpha", "bravo", "charlie"},
		},
		{
			desc:  "AdjacentDuplicates",
			slice: []string{"alpha", "bravo", "bravo", "charlie"},
			want:  []string{"alpha", "bravo", "charlie"},
		},
		{
			desc:  "NonAdjacentDuplicates",
			slice: []string{"alpha", "bravo", "alpha", "charlie"},
			want:  []string{"alpha", "bravo", "charlie"},
		},
		{
			desc:  "MultipleDuplicates",
			slice: []string{"alpha", "bravo", "alpha", "alpha", "charlie", "charlie", "alpha", "bravo"},
			want:  []string{"alpha", "bravo", "charlie"},
		},
		{
			desc:  "UnsortedDuplicates",
			slice: []string{"charlie", "bravo", "alpha", "bravo"},
			want:  []string{"charlie", "bravo", "alpha"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := RemoveDuplicateStrings(tc.slice)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("RemoveDuplicateStrings(%v) = %v, want: %v", tc.slice, got, tc.want)
			}
		})
	}
}

func TestMaskProxyPassword(t *testing.T) {
	type dockerOptTest struct {
		input  string
		output string
	}
	var tests = []dockerOptTest{
		{
			input:  "cats",
			output: "cats",
		},
		{
			input:  "myDockerOption=value",
			output: "myDockerOption=value",
		},
		{
			input:  "http://minikube.sigs.k8s.io",
			output: "http://minikube.sigs.k8s.io",
		},
		{
			input:  "http://jdoe@minikube.sigs.k8s.io:8080",
			output: "http://jdoe@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "https://mary:iam$Fake!password@minikube.sigs.k8s.io:8080",
			output: "https://mary:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http://jdoe:%n0tRe@al:Password!@minikube.sigs.k8s.io:8080",
			output: "http://jdoe:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http://jo@han:n0tRe@al:&Password!@minikube.sigs.k8s.io:8080",
			output: "http://jo@han:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http://k@r3n!:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "http://k@r3n!:*****@minikube.sigs.k8s.io",
		},
		{
			input:  "https://fr@ank5t3in:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "https://fr@ank5t3in:*****@minikube.sigs.k8s.io",
		},
	}
	for _, test := range tests {
		got := MaskProxyPassword(test.input)
		if got != test.output {
			t.Errorf("MaskProxyPassword(\"%v\"): got %v, expected %v", test.input, got, test.output)
		}
	}
}

func TestMaskProxyPasswordWithKey(t *testing.T) {
	type dockerOptTest struct {
		input  string
		output string
	}
	var tests = []dockerOptTest{
		{
			input:  "cats",
			output: "cats",
		},
		{
			input:  "myDockerOption=value",
			output: "myDockerOption=value",
		},
		{
			input:  "http_proxy=http://minikube.sigs.k8s.io",
			output: "HTTP_PROXY=http://minikube.sigs.k8s.io",
		},
		{
			input:  "https_proxy=http://jdoe@minikube.sigs.k8s.io:8080",
			output: "HTTPS_PROXY=http://jdoe@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "https_proxy=https://mary:iam$Fake!password@minikube.sigs.k8s.io:8080",
			output: "HTTPS_PROXY=https://mary:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://jdoe:%n0tRe@al:Password!@minikube.sigs.k8s.io:8080",
			output: "HTTP_PROXY=http://jdoe:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://jo@han:n0tRe@al:&Password!@minikube.sigs.k8s.io:8080",
			output: "HTTP_PROXY=http://jo@han:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://k@r3n!:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "HTTP_PROXY=http://k@r3n!:*****@minikube.sigs.k8s.io",
		},
		{
			input:  "https_proxy=https://fr@ank5t3in:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "HTTPS_PROXY=https://fr@ank5t3in:*****@minikube.sigs.k8s.io",
		},
	}
	for _, test := range tests {
		got := MaskProxyPasswordWithKey(test.input)
		if got != test.output {
			t.Errorf("MaskProxyPasswordWithKey(\"%v\"): got %v, expected %v", test.input, got, test.output)
		}
	}
}
