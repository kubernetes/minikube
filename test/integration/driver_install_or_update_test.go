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
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/blang/semver"

	"k8s.io/minikube/pkg/minikube/driver"
)

func TestKVMDriverInstallOrUpdate(t *testing.T) {
	if NoneDriver() {
		t.Skip("Skip none driver.")
	}

	if runtime.GOOS != "linux" {
		t.Skip("Skip if not linux.")
	}

	MaybeParallel(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "driver-without-version-support", path: filepath.Join(*testdataDir, "kvm2-driver-without-version")},
		{name: "driver-with-older-version", path: filepath.Join(*testdataDir, "kvm2-driver-without-version")},
	}

	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	for _, tc := range tests {
		dir, err := ioutil.TempDir("", tc.name)
		if err != nil {
			t.Fatalf("Expected to create tempdir. test: %s, got: %v", tc.name, err)
		}
		defer func() {
			err := os.RemoveAll(dir)
			if err != nil {
				t.Errorf("Failed to remove dir %q: %v", dir, err)
			}
		}()

		pwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error not expected when getting working directory. test: %s, got: %v", tc.name, err)
		}

		path := filepath.Join(pwd, tc.path)

		_, err = os.Stat(filepath.Join(path, "docker-machine-driver-kvm2"))
		if err != nil {
			t.Fatalf("Expected driver to exist. test: %s, got: %v", tc.name, err)
		}

		// change permission to allow driver to be executable
		err = os.Chmod(filepath.Join(path, "docker-machine-driver-kvm2"), 0700)
		if err != nil {
			t.Fatalf("Expected not expected when changing driver permission. test: %s, got: %v", tc.name, err)
		}

		os.Setenv("PATH", fmt.Sprintf("%s:%s", path, originalPath))

		// NOTE: This should be a real version, as it impacts the downloaded URL
		newerVersion, err := semver.Make("1.3.0")
		if err != nil {
			t.Fatalf("Expected new semver. test: %v, got: %v", tc.name, err)
		}

		err = driver.InstallOrUpdate("kvm2", dir, newerVersion, true, true)
		if err != nil {
			t.Fatalf("Failed to update driver to %v. test: %s, got: %v", newerVersion, tc.name, err)
		}

		_, err = os.Stat(filepath.Join(dir, "docker-machine-driver-kvm2"))
		if err != nil {
			t.Fatalf("Expected driver to be download. test: %s, got: %v", tc.name, err)
		}
	}
}
