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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"k8s.io/minikube/pkg/minikube/localpath"
)

// MakeTempDir creates the temp dir and returns the path
func MakeTempDir() string {
	tempDir, err := ioutil.TempDir("", "minipath")
	if err != nil {
		log.Fatal(err)
	}
	tempDir = filepath.Join(tempDir, ".minikube")
	err = os.MkdirAll(filepath.Join(tempDir, "addons"), 0o777)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(filepath.Join(tempDir, "cache", "iso"), 0o777)
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv(localpath.MinikubeHome, tempDir)
	return localpath.MiniPath()
}

func RemoveTempDir(tempdir string) {
	if filepath.Base(tempdir) == ".minikube" {
		tempdir = filepath.Dir(tempdir)
	}
	os.RemoveAll(tempdir)
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
