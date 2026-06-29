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

	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
)

// TestHelmInstall integration test verifies helm installation, upgrade, and no-change behavior inside a live guest VM.
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

	// 1. Install test
	t.Run("Install", func(t *testing.T) {
		_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm", "/usr/local/bin/helm"))
		if err != nil {
			t.Fatalf("failed to delete helm inside guest: %v", err)
		}

		err = addons.HelmInstallBinary(nil, runner)
		if err != nil {
			t.Fatalf("HelmInstallBinary failed: %v", err)
		}

		_, err = runner.RunCmd(exec.Command("test", "-f", "/usr/bin/helm"))
		if err != nil {
			t.Errorf("helm binary not found at /usr/bin/helm after install: %v", err)
		}
	})

	// 2. Upgrade test
	t.Run("Upgrade", func(t *testing.T) {
		_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm"))
		if err != nil {
			t.Fatalf("failed to delete /usr/bin/helm: %v", err)
		}

		_, err = runner.RunCmd(exec.Command("sudo", "bash", "-c", "echo 'echo old-helm' > /usr/local/bin/helm && chmod +x /usr/local/bin/helm"))
		if err != nil {
			t.Fatalf("failed to place old helm script in /usr/local/bin/helm: %v", err)
		}

		err = addons.HelmInstallBinary(nil, runner)
		if err != nil {
			t.Fatalf("HelmInstallBinary failed: %v", err)
		}

		_, err = runner.RunCmd(exec.Command("test", "-f", "/usr/bin/helm"))
		if err != nil {
			t.Errorf("helm binary not found at /usr/bin/helm after upgrade: %v", err)
		}

		rr, err := runner.RunCmd(exec.Command("/usr/bin/helm", "version"))
		if err != nil {
			t.Errorf("helm version command failed after upgrade: %v", err)
		}
		if strings.Contains(rr.Stdout.String(), "old-helm") {
			t.Errorf("helm was not successfully upgraded/replaced: stdout: %s", rr.Stdout.String())
		}
	})

	// 3. No Change test
	t.Run("NoChange", func(t *testing.T) {
		err = addons.HelmInstallBinary(nil, runner)
		if err != nil {
			t.Fatalf("HelmInstallBinary failed: %v", err)
		}

		_, err = runner.RunCmd(exec.Command("test", "-f", "/usr/bin/helm"))
		if err != nil {
			t.Errorf("helm binary not found at /usr/bin/helm after second execution: %v", err)
		}
	})
}
