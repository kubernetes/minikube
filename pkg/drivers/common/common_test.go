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

package common

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/tests"
)

func TestCreateRawDiskImage(t *testing.T) {
	tmpdir := tests.MakeTempDir(t)

	sshPath := filepath.Join(tmpdir, "ssh")
	if err := os.WriteFile(sshPath, []byte("mysshkey"), 0644); err != nil {
		t.Fatalf("writefile: %v", err)
	}
	diskPath := filepath.Join(tmpdir, "disk")

	sizeInMb := 100
	sizeInBytes := int64(104857600)
	if err := CreateRawDiskImage(sshPath, diskPath, sizeInMb); err != nil {
		t.Fatalf("CreateRawDiskImage() error = %v", err)
	}
	fi, err := os.Lstat(diskPath)
	if err != nil {
		t.Fatalf("Lstat() error = %v", err)
	}
	if fi.Size() != sizeInBytes {
		t.Errorf("Disk size is %v, want %v", fi.Size(), sizeInBytes)
	}
	f, err := os.Open(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	tr := tar.NewReader(f)
	for _, want := range []struct {
		name string
		body string
		mode int64
	}{
		{"boot2docker, please format-me", "boot2docker, please format-me", 0},
		{".ssh", "", 0700},
		{".ssh/authorized_keys", "mysshkey", 0644},
		{".ssh/authorized_keys2", "mysshkey", 0644},
	} {
		h, err := tr.Next()
		if err != nil {
			t.Fatal(err)
		}
		if h.Name != want.name || h.Mode != want.mode {
			t.Errorf("tar entry = %s (%o), want %s (%o)", h.Name, h.Mode, want.name, want.mode)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != want.body {
			t.Errorf("%s content = %q, want %q", h.Name, body, want.body)
		}
	}
	if _, err := tr.Next(); err != io.EOF {
		t.Fatalf("after final tar entry: got %v, want EOF", err)
	}
}

func TestCreateRawDiskImageErrors(t *testing.T) {
	for _, scenario := range []string{"missing key", "missing directory", "existing disk", "invalid size"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			keyPath := filepath.Join(dir, "key.pub")
			diskPath := filepath.Join(dir, "disk")
			if err := os.WriteFile(keyPath, []byte("test public key"), 0644); err != nil {
				t.Fatal(err)
			}
			size := 1
			switch scenario {
			case "missing key":
				keyPath = filepath.Join(dir, "missing.pub")
			case "missing directory":
				diskPath = filepath.Join(dir, "missing", "disk")
			case "existing disk":
				if err := os.WriteFile(diskPath, []byte("existing disk contents"), 0644); err != nil {
					t.Fatal(err)
				}
			case "invalid size":
				size = -1
			}
			err := CreateRawDiskImage(keyPath, diskPath, size)
			var pathErr *os.PathError
			if !errors.As(err, &pathErr) {
				t.Fatalf("CreateRawDiskImage() error = %v, want filesystem error", err)
			}
			if scenario == "existing disk" {
				data, err := os.ReadFile(diskPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != "existing disk contents" {
					t.Fatalf("existing disk was modified: %q", data)
				}
			}
		})
	}
}
