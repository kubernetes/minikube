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
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs"
	"k8s.io/minikube/pkg/minikube/run"
)

// minExpectedHelmVersion is the minimum helm version we expect to be installed.
// We cannot check for the exact latest version because during Helm release rollouts,
// CDN propagation delays can cause our version query and the get_helm.sh install script
// to see different versions, leading to flaky test failures.
// See https://github.com/kubernetes/minikube/issues/23323
// Bumped occasionally (not on every release) to catch the CDN serving something
// wildly stale or broken, without needing to track the exact current latest.
var minExpectedHelmVersion = semver.Version{Major: 3, Minor: 20}

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
		t.Logf("installed helm version: %q", version)
		parsed, err := semver.Parse(strings.TrimPrefix(version, "v"))
		if err != nil {
			t.Fatalf("failed to parse helm version: %q: %v", version, err)
		}
		if parsed.LT(minExpectedHelmVersion) {
			t.Fatalf("installed helm version %q is older than minimum expected %q", parsed, minExpectedHelmVersion)
		}
	})

	// 2. Upgrade test
	// This test ensures that if minExpectedHelmVersion is installed at /usr/bin/helm,
	// calling InstallHelm with HelmOptions{} (defaults to latest) will correctly upgrade it to the latest version.
	t.Run("Upgrade", func(t *testing.T) {
		// Install minExpectedHelmVersion directly to /usr/bin/helm
		err := addons.InstallHelm(runner, addons.HelmOptions{Version: minExpectedHelmVersion.String()})
		if err != nil {
			t.Fatalf("failed to install helm %s: %v", minExpectedHelmVersion, err)
		}

		// Verify the pinned helm version was installed and is executable
		versionOld := addons.HelmVersion(runner)
		t.Logf("installed helm version: %s (expected %s)", versionOld, minExpectedHelmVersion)
		parsedOld, err := semver.Parse(strings.TrimPrefix(versionOld, "v"))
		if err != nil {
			t.Fatalf("failed to parse helm version: %q: %v", versionOld, err)
		}
		if parsedOld.NE(minExpectedHelmVersion) {
			t.Fatalf("helm version mismatch: expected %q, got %q", minExpectedHelmVersion, versionOld)
		}

		// Run InstallHelm, which should download and install the latest helm into /usr/bin/helm
		err = addons.InstallHelm(runner, addons.HelmOptions{})
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		// Verify that a newer version of helm has been installed in /usr/bin/helm
		versionNew := addons.HelmVersion(runner)
		t.Logf("upgraded helm version: %s (from %s)", versionNew, versionOld)
		parsedNew, err := semver.Parse(strings.TrimPrefix(versionNew, "v"))
		if err != nil {
			t.Fatalf("failed to parse helm version: %q: %v", versionNew, err)
		}
		if parsedNew.LTE(minExpectedHelmVersion) {
			t.Fatalf("installed version %q not newer than older version %q", parsedNew, minExpectedHelmVersion)
		}
	})

	// 3. No Change test
	// This test verifies that if Helm is already installed in /usr/bin/helm,
	// running InstallHelm again is a no-op and does not modify or reinstall it.
	t.Run("NoChange", func(t *testing.T) {
		// Run InstallHelm to ensure the pinned helm version is present
		err := addons.InstallHelm(runner, addons.HelmOptions{Version: minExpectedHelmVersion.String()})
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		// Retrieve current helm version details
		firstVersion := addons.HelmVersion(runner)
		t.Logf("first helm version: %s", firstVersion)
		if firstVersion == "" {
			t.Fatalf("helm not found at /usr/bin/helm")
		}

		// Run InstallHelm again with the same pinned version
		err = addons.InstallHelm(runner, addons.HelmOptions{Version: minExpectedHelmVersion.String()})
		if err != nil {
			t.Fatalf("InstallHelm second call failed: %v", err)
		}

		// Retrieve helm version details again and compare
		secondVersion := addons.HelmVersion(runner)
		t.Logf("second helm version: %s", secondVersion)
		if secondVersion == "" {
			t.Fatalf("helm not found at /usr/bin/helm")
		}
		if firstVersion != secondVersion {
			t.Fatalf("helm version changed after second InstallHelm call: first %q, second %q", firstVersion, secondVersion)
		}
	})
}
