/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/blang/semver/v4"

	"k8s.io/minikube/pkg/minikube/driver/auxdriver"
)

// TestKVMDriverInstallOrUpdate makes sure our docker-machine-driver-kvm2 binary can be installed properly
func TestKVMDriverInstallOrUpdate(t *testing.T) {
	if NoneDriver() {
		t.Skip("Skip none driver.")
	}

	if runtime.GOOS != "linux" {
		t.Skip("Skip if not linux.")
	}

	if arm64Platform() {
		t.Skip("Skip if arm64. See https://github.com/kubernetes/minikube/issues/10144")
	}

	MaybeParallel(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "driver-with-older-version", path: filepath.Join(*testdataDir, "kvm2-driver-older-version")},
	}

	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	for _, tc := range tests {
		tempDLDir := t.TempDir()

		pwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error not expected when getting working directory. test: %s, got: %v", tc.name, err)
		}

		path := filepath.Join(pwd, tc.path)

		_, err = os.Stat(filepath.Join(path, "docker-machine-driver-kvm2"))
		if err != nil {
			t.Fatalf("Expected test data driver to exist. test: %s, got: %v", tc.name, err)
		}

		// copy test data driver into the temp download dir so we can point PATH to it for before/after install
		src := filepath.Join(path, "docker-machine-driver-kvm2")
		dst := filepath.Join(tempDLDir, "docker-machine-driver-kvm2")
		if err = CopyFile(src, dst, false); err != nil {
			t.Fatalf("Failed to copy test data driver to temp dir. test: %s, got: %v", tc.name, err)
		}

		// point to the copied driver for the rest of the test
		path = tempDLDir

		// change permission to allow driver to be executable
		err = os.Chmod(filepath.Join(path, "docker-machine-driver-kvm2"), 0700)
		if err != nil {
			t.Fatalf("Expected not expected when changing driver permission. test: %s, got: %v", tc.name, err)
		}

		os.Setenv("PATH", fmt.Sprintf("%s:%s", path, originalPath))

		// NOTE: This should be a real version, as it impacts the downloaded URL
		newerVersion, err := semver.Make("1.37.0")
		if err != nil {
			t.Fatalf("Expected new semver. test: %v, got: %v", tc.name, err)
		}

		err = auxdriver.InstallOrUpdate("kvm2", tempDLDir, newerVersion, true, true)
		if err != nil {
			t.Fatalf("Failed to update driver to %v. test: %s, got: %v", newerVersion, tc.name, err)
		}

		_, err = os.Stat(filepath.Join(tempDLDir, "docker-machine-driver-kvm2"))
		if err != nil {
			t.Fatalf("Expected driver to be download. test: %s, got: %v", tc.name, err)
		}
	}
}
