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

package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/localpath"
)

// MakeTempDir creates the temp dir and returns the path
func MakeTempDir(t *testing.T) string {
	tempDir := t.TempDir()
	tempDir = filepath.Join(tempDir, ".minikube")
	if err := os.MkdirAll(filepath.Join(tempDir, "addons"), 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "cache"), 0777); err != nil {
		t.Fatal(err)
	}
	os.Setenv(localpath.MinikubeHome, tempDir)
	return localpath.MiniPath()
}

// FakeFile satisfies fdWriter
type FakeFile struct {
	b bytes.Buffer
}

// NewFakeFile creates a FakeFile
func NewFakeFile() *FakeFile {
	return &FakeFile{}
}

// Fd returns the file descriptor
func (f *FakeFile) Fd() uintptr {
	return uintptr(0)
}

func (f *FakeFile) Write(p []byte) (int, error) {
	return f.b.Write(p)
}
func (f *FakeFile) String() string {
	return f.b.String()
}
