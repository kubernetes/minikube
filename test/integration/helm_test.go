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
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs"
	"k8s.io/minikube/pkg/minikube/run"
)

// TestHelmInstall integration test verifies helm installation, upgrade, and no-change behavior inside a live guest VM.
// We test this flow because installing the latest helm allows self-healing and updating the cluster when Helm is missing, outdated, or broken.
func TestHelmInstall(t *testing.T) {
	MaybeParallel(t)

	profile := UniqueProfileName("helm")
	// TODO: update the timeout based on data from recent runs (3*p95 once data is available)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := append([]string{"start", "-p", profile, "--memory=3072", "--alsologtostderr"}, StartArgs()...)
	_, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start minikube: %v", err)
	}

	cc, err := config.Load(profile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	api, err := machine.NewAPIClient(&run.CommandOptions{NonInteractive: true})
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

	// Check that the guest has outbound internet access by probing the Helm download endpoint.
	// On Prow CI with docker driver, the guest container may not have outbound connectivity.
	_, err = runner.RunCmd(exec.Command("curl", "-fsSL", "--max-time", "10", "-o", "/dev/null", "https://get.helm.sh/helm3-latest-version"))
	if err != nil {
		t.Skip("skipping: guest VM/container has no outbound internet access (this is required to download helm): https://github.com/kubernetes/minikube/issues/23275")
	}

	latestVersion := getLatestHelmVersion(t)

	// 1. Install test
	// The minikube ISO and kicbase images already come with Helm pre-installed.
	// We delete the Helm binary here to simulate a broken cluster (e.g. if a user manually deleted or corrupted it).
	// This test ensures that installing the latest helm allows self-healing and recovering when Helm is missing or broken.
	t.Run("Install", func(t *testing.T) {
		_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm"))
		if err != nil {
			t.Fatalf("failed to delete helm inside guest: %v", err)
		}

		err = addons.InstallHelm(runner, addons.HelmOptions{})
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		version := addons.HelmVersion(runner)
		if version != latestVersion {
			t.Fatalf("installed helm version mismatch: expected %q, got %q", latestVersion, version)
		}
	})

	// 2. Upgrade test
	// This test ensures that if an older version of Helm is installed at /usr/bin/helm,
	// calling InstallHelm with HelmOptions{} (defaults to latest) will correctly upgrade it to the latest version.
	t.Run("Upgrade", func(t *testing.T) {
		// Install an older Helm version (e.g. v3.12.0) directly to /usr/bin/helm
		err := addons.InstallHelm(runner, addons.HelmOptions{Version: "v3.12.0"})
		if err != nil {
			t.Fatalf("failed to install older helm: %v", err)
		}

		// Verify the older helm was installed and is executable
		versionOld := addons.HelmVersion(runner)
		if versionOld != "v3.12.0" {
			t.Fatalf("older helm version mismatch: expected \"v3.12.0\", got %q", versionOld)
		}

		// Run InstallHelm, which should download and install the latest helm into /usr/bin/helm
		err = addons.InstallHelm(runner, addons.HelmOptions{})
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		// Verify that a newer version of helm has been installed in /usr/bin/helm
		versionNew := addons.HelmVersion(runner)
		if versionNew == versionOld {
			t.Fatalf("helm version was not upgraded: still %q", versionNew)
		}
		if versionNew != latestVersion {
			t.Fatalf("upgraded helm version mismatch: expected %q, got %q", latestVersion, versionNew)
		}
	})

	// 3. No Change test
	// This test verifies that if Helm is already installed in /usr/bin/helm,
	// running InstallHelm again is a no-op and does not modify or reinstall it.
	t.Run("NoChange", func(t *testing.T) {
		// Run InstallHelm to ensure latest helm is present
		err := addons.InstallHelm(runner, addons.HelmOptions{})
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		// Retrieve current helm version details
		firstVersion := addons.HelmVersion(runner)
		if firstVersion == "" {
			t.Fatalf("helm not found at /usr/bin/helm")
		}

		// Run InstallHelm again
		err = addons.InstallHelm(runner, addons.HelmOptions{})
		if err != nil {
			t.Fatalf("InstallHelm second call failed: %v", err)
		}

		// Retrieve helm version details again and compare
		secondVersion := addons.HelmVersion(runner)
		if secondVersion == "" {
			t.Fatalf("helm not found at /usr/bin/helm")
		}
		if firstVersion != secondVersion {
			t.Fatalf("helm version changed after second InstallHelm call: first %q, second %q", firstVersion, secondVersion)
		}
	})
}

// getLatestHelmVersion queries the official Helm 3 latest version URL (used by get_helm.sh)
// to find the latest Helm 3 release tag.
func getLatestHelmVersion(t *testing.T) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://get.helm.sh/helm3-latest-version", nil)
	if err != nil {
		t.Fatalf("failed to create request for latest helm version: %v", err)
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to query latest helm release tag: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code querying latest helm release tag: %d", resp.StatusCode)
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
		return ""
	}
	return strings.TrimSpace(string(body))
}
