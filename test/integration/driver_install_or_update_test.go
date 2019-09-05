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
	"path"
	"runtime"
	"testing"

	"github.com/blang/semver"

	"k8s.io/minikube/pkg/drivers"
)

func TestDriverInstallOrUpdate(t *testing.T) {
	if isTestNoneDriver(t) {
		t.Skip("Skip none driver.")
	}

	if runtime.GOOS != "linux" {
		t.Skip("Skip if not linux.")
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "driver-without-version-support", path: path.Join(*testdataDir, "kvm2-driver-without-version")},
		{name: "driver-with-older-version", path: path.Join(*testdataDir, "kvm2-driver-without-version")},
	}

	for _, tc := range tests {
		dir, err := ioutil.TempDir("", tc.name)
		if err != nil {
			t.Fatalf("Expected to create tempdir. test: %s, got: %v", tc.name, err)
		}
		defer os.RemoveAll(dir)

		os.Setenv("PATH", fmt.Sprintf("%s:%s", tc.path, os.Getenv("PATH")))

		newerVersion, err := semver.Make("1.1.3")
		if err != nil {
			t.Fatalf("Expected new semver. test: %v, got: %v", tc.name, err)
		}

		err = drivers.InstallOrUpdate("docker-machine-driver-kvm2", dir, newerVersion)
		if err != nil {
			t.Fatalf("Expected to update driver. test: %s, got: %v", tc.name, err)
		}

		_, err = os.Stat(path.Join(dir, "docker-machine-driver-kvm2"))
		if err != nil {
			t.Fatalf("Expected driver to be download. test: %s, got: %v", tc.name, err)
		}
	}
}
