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
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
)

func TestGetISOFileURI(t *testing.T) {
	dler := DefaultDownloader{}

	tests := map[string]string{
		"file:///test/path/minikube-test.iso":                           "file:///test/path/minikube-test.iso",
		"https://storage.googleapis.com/minikube/iso/minikube-test.iso": "file://" + filepath.ToSlash(filepath.Join(constants.GetMinipath(), "cache", "iso", "minikube-test.iso")),
	}

	for input, expected := range tests {
		if isoFileURI := dler.GetISOFileURI(input); isoFileURI != expected {
			t.Fatalf("Expected GetISOFileURI with input %s to return %s but instead got: %s", input, expected, isoFileURI)
		}
	}

}

var testISOString = "hello"

func TestCacheMinikubeISOFromURL(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)
	dler := DefaultDownloader{}
	isoPath := filepath.Join(constants.GetMinipath(), "cache", "iso", "minikube-test.iso")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, testISOString); err != nil {
			t.Fatalf("WriteString: %v", err)
		}
	}))
	isoURL := server.URL + "/minikube-test.iso"
	if err := dler.CacheMinikubeISOFromURL(isoURL); err != nil {
		t.Fatalf("Unexpected error from CacheMinikubeISOFromURL: %v", err)
	}

	transferred, err := ioutil.ReadFile(filepath.Join(isoPath))
	if err != nil {
		t.Fatalf("File not copied. Could not open file at path: %s", isoPath)
	}

	//test that the ISO is transferred properly
	contents := []byte(testISOString)
	if !bytes.Contains(transferred, contents) {
		t.Fatalf("Expected transfers to contain: %s. It was: %s", contents, transferred)
	}

}

func TestShouldCacheMinikubeISO(t *testing.T) {
	dler := DefaultDownloader{}

	tests := map[string]bool{
		"file:///test/path/minikube-test.iso":                           false,
		"https://storage.googleapis.com/minikube/iso/minikube-test.iso": true,
	}

	for input, expected := range tests {
		if out := dler.ShouldCacheMinikubeISO(input); out != expected {
			t.Fatalf("Expected ShouldCacheMinikubeISO with input %s to return %t but instead got: %t", input, expected, out)
		}
	}
}

func TestIsMinikubeISOCached(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	dler := DefaultDownloader{}

	testFileURI := "file:///test/path/minikube-test.iso"
	expected := false

	if out := dler.IsMinikubeISOCached(testFileURI); out != expected {
		t.Fatalf("Expected IsMinikubeISOCached with input %s to return %t but instead got: %t", testFileURI, expected, out)
	}

	if err := ioutil.WriteFile(filepath.Join(constants.GetMinipath(), "cache", "iso", "minikube-test.iso"), []byte(testISOString), os.FileMode(int(0644))); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	expected = true
	if out := dler.IsMinikubeISOCached(testFileURI); out != expected {
		t.Fatalf("Expected IsMinikubeISOCached with input %s to return %t but instead got: %t", testFileURI, expected, out)
	}

}
