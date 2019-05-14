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

package drivers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/tests"
)

func Test_createDiskImage(t *testing.T) {
	tmpdir := tests.MakeTempDir()
	defer os.RemoveAll(tmpdir)

	sshPath := filepath.Join(tmpdir, "ssh")
	if err := ioutil.WriteFile(sshPath, []byte("mysshkey"), 0644); err != nil {
		t.Fatalf("writefile: %v", err)
	}
	diskPath := filepath.Join(tmpdir, "disk")

	sizeInMb := 100
	sizeInBytes := int64(sizeInMb) * 1000000
	if err := createRawDiskImage(sshPath, diskPath, sizeInMb); err != nil {
		t.Errorf("createDiskImage() error = %v", err)
	}
	fi, err := os.Lstat(diskPath)
	if err != nil {
		t.Errorf("Lstat() error = %v", err)
	}
	if fi.Size() != sizeInBytes {
		t.Errorf("Disk size is %v, want %v", fi.Size(), sizeInBytes)
	}
}
