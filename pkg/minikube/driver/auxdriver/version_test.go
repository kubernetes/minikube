/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package auxdriver

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/blang/semver/v4"
	"k8s.io/minikube/pkg/minikube/driver"
)

func TestMinAcceptableDriverVersion(t *testing.T) {
	tests := []struct {
		desc            string
		driver          string
		minikubeVersion string
		wantedVersion   semver.Version
	}{
		{"Hyperkit", driver.HyperKit, "1.1.1", *minHyperkitVersion},
		{"Invalid", "_invalid_", "1.1.1", semanticVersion("1.1.1")},
		{"KVM2", driver.KVM2, "1.1.1", semanticVersion("1.1.1")},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := minAcceptableDriverVersion(tt.driver, semanticVersion(tt.minikubeVersion)); !got.EQ(tt.wantedVersion) {
				t.Errorf("Invalid min acceptable driver version, got: %v, want: %v", got, tt.wantedVersion)
			}
		})
	}
}

func semanticVersion(s string) semver.Version {
	r, err := semver.New(s)
	if err != nil {
		panic(err)
	}
	return *r
}

func TestDriverVersion(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		path := buildTestDriver(t, "valid")
		v, err := driverVersion(path)
		if err != nil {
			t.Fatalf("failed to get driver version: %s", err)
		}
		expected := Version{Version: "v1.2.3", Commit: "1af8bdc072232de4b1fec3b6cc0e8337e118bc83"}
		if v != expected {
			t.Errorf("Invalid driver version, got: %v, want: %v", v, expected)
		}
	})

	t.Run("no version", func(t *testing.T) {
		path := buildTestDriver(t, "no-version")
		if _, err := driverVersion(path); err == nil {
			t.Fatalf("missing version did not fail")
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("no commit", func(t *testing.T) {
		path := buildTestDriver(t, "no-commit")
		if _, err := driverVersion(path); err == nil {
			t.Fatalf("missing commit did not fail")
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		path := buildTestDriver(t, "invalid")
		if _, err := driverVersion(path); err == nil {
			t.Fatalf("invalid yaml did not fail")
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("fail", func(t *testing.T) {
		path := buildTestDriver(t, "fail")
		if _, err := driverVersion(path); err == nil {
			t.Fatalf("failing driver did not fail")
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "driver.exe")
		if _, err := driverVersion(path); err == nil {
			t.Fatalf("missing driver did not fail")
		} else {
			t.Logf("expected error: %v", err)
		}
	})
}

func buildTestDriver(t *testing.T, name string) string {
	out := filepath.Join(t.TempDir(), name)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", out, "testdata/driver.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	t.Logf("Building %q", out)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	return out
}
