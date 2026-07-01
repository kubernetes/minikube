//go:build integration

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

package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs"
)

// TestHelmInstall integration test verifies helm installation, upgrade, and no-change behavior inside a live guest VM.
// We test this flow because installing the latest helm allows self-healing and updating the cluster when Helm is missing, outdated, or broken.
func TestHelmInstall(t *testing.T) {
	MaybeParallel(t)

	profile := UniqueProfileName("helm")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := append([]string{"start", "-p", profile, "--memory=2048", "--alsologtostderr", "-v=1"}, StartArgs()...)
	_, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start minikube: %v", err)
	}

	cc, err := config.Load(profile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	api, err := machine.NewAPIClient(nil)
	if err != nil {
		t.Fatalf("failed to create api client: %v", err)
	}
	defer api.Close()

	cp, err := config.ControlPlane(*cc)
	if err != nil {
		t.Fatalf("failed to get control plane node: %v", err)
	}

	host, err := machine.LoadHost(api, config.MachineName(*cc, cp))
	if err != nil {
		t.Fatalf("failed to load host: %v", err)
	}

	runner, err := machine.CommandRunner(host)
	if err != nil {
		t.Fatalf("failed to get command runner: %v", err)
	}

	latestVersion, errLatest := getLatestHelmVersion()
	if errLatest != nil {
		t.Logf("warning: failed to fetch latest helm version from github: %v (falling back to simple version checks)", errLatest)
	}

	// 1. Install test
	// The minikube ISO and kicbase images already come with Helm pre-installed.
	// We delete the Helm binary here to simulate a broken cluster (e.g. if a user manually deleted or corrupted it).
	// This test ensures that installing the latest helm allows self-healing and recovering when Helm is missing or broken.
	t.Run("Install", func(t *testing.T) {
		_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm", "/usr/local/bin/helm"))
		if err != nil {
			t.Fatalf("failed to delete helm inside guest: %v", err)
		}

		err = addons.InstallHelm(nil, runner, "latest")
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		verifyHelmVersion(t, runner, "/usr/bin/helm", latestVersion)
	})

	// 2. Upgrade test
	// This test ensures that if /usr/bin/helm is missing, but an older version of Helm is
	// installed at /usr/local/bin/helm, InstallHelm will correctly download and install
	// the latest version of Helm directly to /usr/bin/helm.
	t.Run("Upgrade", func(t *testing.T) {
		t.Cleanup(func() {
			_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/local/bin/helm"))
			if err != nil {
				t.Logf("warning: failed to delete temporary older helm: %v", err)
			}
		})

		// Ensure /usr/bin/helm is deleted so the installer triggers
		_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm"))
		if err != nil {
			t.Fatalf("failed to delete /usr/bin/helm: %v", err)
		}

		// Install an older Helm version (e.g. v3.12.0) to /usr/local/bin/helm
		err = addons.InstallHelmVersion(runner, "v3.12.0", "/usr/local/bin")
		if err != nil {
			t.Fatalf("failed to install older helm: %v", err)
		}

		// Verify the older helm was installed and is executable
		verifyHelmVersion(t, runner, "/usr/local/bin/helm", "v3.12.0")

		// Run InstallHelm, which should download and install the latest helm into /usr/bin/helm
		err = addons.InstallHelm(nil, runner, "latest")
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		// Verify that a newer version of helm has been installed in /usr/bin/helm
		versionNew := verifyHelmVersion(t, runner, "/usr/bin/helm", latestVersion)
		if strings.Contains(versionNew, "v3.12.0") {
			t.Errorf("helm was not successfully upgraded/replaced: got %s", versionNew)
		}
	})

	// 3. No Change test
	// This test verifies that if Helm is already installed in /usr/bin/helm,
	// running InstallHelm again is a no-op and does not modify or reinstall it.
	t.Run("NoChange", func(t *testing.T) {
		// Run InstallHelm to ensure latest helm is present
		err = addons.InstallHelm(nil, runner, "latest")
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		// Retrieve current helm version details
		firstVersion := verifyHelmVersion(t, runner, "/usr/bin/helm", "")

		// Run InstallHelm again
		err = addons.InstallHelm(nil, runner, "latest")
		if err != nil {
			t.Fatalf("InstallHelm second call failed: %v", err)
		}

		// Retrieve helm version details again and compare
		verifyHelmVersion(t, runner, "/usr/bin/helm", firstVersion)
	})
}

// verifyHelmVersion runs 'helm version --template {{.Version}}' inside the guest
// and asserts it matches the expected substring if expectedVersion is not empty.
func verifyHelmVersion(t *testing.T, runner command.Runner, path string, expectedVersion string) string {
	t.Helper()
	cmd := exec.Command(path, "version", "--template", "{{.Version}}")
	rr, err := runner.RunCmd(cmd)
	if err != nil {
		t.Fatalf("failed to check helm version at %s: %v", path, err)
	}
	version := strings.TrimSpace(rr.Stdout.String())
	if expectedVersion != "" && !strings.Contains(version, expectedVersion) {
		t.Fatalf("helm version mismatch at %s: expected %q, got %q", path, expectedVersion, version)
	}
	return version
}

// getLatestHelmVersion queries the official Helm 3 latest version URL (used by get_helm.sh)
// to find the latest Helm 3 release tag.
func getLatestHelmVersion() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get("https://get.helm.sh/helm3-latest-version")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code querying latest release tag: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}
