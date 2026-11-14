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

package qemu

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/libmachine/drivers"
)

func TestGenerateDiskImageWriteError(t *testing.T) {
	d := &Driver{BaseDriver: &drivers.BaseDriver{
		MachineName: "test",
		StorePath:   t.TempDir(),
	}}
	if err := os.MkdirAll(filepath.Dir(d.publicSSHKeyPath()), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(d.publicSSHKeyPath(), []byte("test public key"), 0644); err != nil {
		t.Fatal(err)
	}
	// A directory at the raw image path makes writing fail on every platform,
	// including when the test runs with elevated privileges.
	rawPath := d.diskPath() + ".raw"
	if err := os.Mkdir(rawPath, 0755); err != nil {
		t.Fatal(err)
	}
	err := d.generateDiskImage(1)
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) || pathErr.Path != rawPath {
		t.Fatalf("generateDiskImage() error = %v, want filesystem error for %s", err, rawPath)
	}
}
