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
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Azure/azure-sdk-for-go/tools/apidiff/ioext"
	"github.com/blang/semver"

	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/version"
)

func TestKVMDriverInstallOrUpdate(t *testing.T) {
	if NoneDriver() {
		t.Skip("Skip none driver.")
	}

	if runtime.GOOS != "linux" {
		t.Skip("Skip if not linux.")
	}

	if arm64Platform() {
		t.Skip("Skip if arm64.")
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

func TestHyperKitDriverInstallOrUpdate(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skip if not darwin.")
	}

	MaybeParallel(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "driver-without-version-support", path: filepath.Join(*testdataDir, "hyperkit-driver-without-version")},
		{name: "driver-with-older-version", path: filepath.Join(*testdataDir, "hyperkit-driver-without-version")},
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

		_, err = os.Stat(filepath.Join(path, "docker-machine-driver-hyperkit"))
		if err != nil {
			t.Fatalf("Expected driver to exist. test: %s, got: %v", tc.name, err)
		}

		// change permission to allow driver to be executable
		err = os.Chmod(filepath.Join(path, "docker-machine-driver-hyperkit"), 0700)
		if err != nil {
			t.Fatalf("Expected not expected when changing driver permission. test: %s, got: %v", tc.name, err)
		}

		os.Setenv("PATH", fmt.Sprintf("%s:%s", path, originalPath))

		// NOTE: This should be a real version, as it impacts the downloaded URL
		newerVersion, err := semver.Make("1.3.0")
		if err != nil {
			t.Fatalf("Expected new semver. test: %v, got: %v", tc.name, err)
		}

		if sudoNeedsPassword() {
			t.Skipf("password required to execute 'sudo', skipping remaining test")
		}

		err = driver.InstallOrUpdate("hyperkit", dir, newerVersion, false, true)
		if err != nil {
			t.Fatalf("Failed to update driver to %v. test: %s, got: %v", newerVersion, tc.name, err)
		}

		_, err = os.Stat(filepath.Join(dir, "docker-machine-driver-hyperkit"))
		if err != nil {
			t.Fatalf("Expected driver to be download. test: %s, got: %v", tc.name, err)
		}
	}
}

func TestHyperkitDriverSkipUpgrade(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skip if not darwin.")
	}

	MaybeParallel(t)
	tests := []struct {
		name            string
		path            string
		expectedVersion string
	}{
		{
			name:            "upgrade-v1.11.0-to-current",
			path:            filepath.Join(*testdataDir, "hyperkit-driver-version-1.11.0"),
			expectedVersion: "v1.11.0",
		},
		{
			name:            "upgrade-v1.2.0-to-current",
			path:            filepath.Join(*testdataDir, "hyperkit-driver-older-version"),
			expectedVersion: version.GetVersion(),
		},
	}

	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		t.Fatalf("No sudo in path: %v", err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mkDir, drvPath, err := prepareTempMinikubeDirWithHyperkitDriver(tc.name, tc.path)
			if err != nil {
				t.Fatalf("Failed to prepare tempdir. test: %s, got: %v", tc.name, err)
			}
			defer func() {
				if err := os.RemoveAll(mkDir); err != nil {
					t.Errorf("Failed to remove mkDir %q: %v", mkDir, err)
				}
			}()

			cmd := exec.Command(Target(), "start", "--download-only", "--interactive=false", "--driver=hyperkit")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PATH=%v%c%v", filepath.Dir(drvPath), filepath.ListSeparator, filepath.Dir(sudoPath)),
				"MINIKUBE_HOME="+mkDir)
			if err = cmd.Run(); err != nil {
				t.Fatalf("failed to run minikube. got: %v", err)
			}

			upgradedVersion, err := driverVersion(drvPath)
			if err != nil {
				t.Fatalf("failed to check driver version. got: %v", err)
			}

			if upgradedVersion != tc.expectedVersion {
				t.Fatalf("invalid driver version. expected: %v, got: %v", tc.expectedVersion, upgradedVersion)
			}
		})
	}
}

func sudoNeedsPassword() bool {
	err := exec.Command("sudo", "-n", "ls").Run()
	return err != nil
}

func driverVersion(path string) (string, error) {
	output, err := exec.Command(path, "version").Output()
	if err != nil {
		return "", err
	}

	var resultVersion string
	_, err = fmt.Sscanf(string(output), "version: %s\n", &resultVersion)
	if err != nil {
		return "", err
	}
	return resultVersion, nil
}

// prepareTempMinikubeDirWithHyperkitDriver creates a temp .minikube directory
// with structure essential to testing of hyperkit driver updates
func prepareTempMinikubeDirWithHyperkitDriver(name, driver string) (string, string, error) {
	temp, err := ioutil.TempDir("", name)
	if err != nil {
		return "", "", fmt.Errorf("failed to create tempdir: %v", err)
	}
	mkDir := filepath.Join(temp, ".minikube")
	mkBinDir := filepath.Join(mkDir, "bin")
	err = os.MkdirAll(mkBinDir, 0777)
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare tempdir: %v", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get working directory: %v", err)
	}

	testDataDriverPath := filepath.Join(pwd, driver, "docker-machine-driver-hyperkit")
	if _, err = os.Stat(testDataDriverPath); err != nil {
		return "", "", fmt.Errorf("expected driver to exist: %v", err)
	}
	// copy driver to temp bin
	testDriverPath := filepath.Join(mkBinDir, "docker-machine-driver-hyperkit")
	if err = ioext.CopyFile(testDataDriverPath, testDriverPath, false); err != nil {
		return "", "", fmt.Errorf("failed to setup current hyperkit driver: %v", err)
	}

	// try to copy cached files to the temp minikube folder to avoid downloading of iso and preloads
	_ = ioext.CopyDir(filepath.Join(localpath.MakeMiniPath("cache")), filepath.Join(mkDir, "cache"))

	// change permission to allow driver to be executable
	if err = os.Chmod(testDriverPath, 0755); err != nil {
		return "", "", fmt.Errorf("failed to set driver permission: %v", err)
	}
	return temp, testDriverPath, nil
}
