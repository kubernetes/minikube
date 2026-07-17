/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package addons

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/blang/semver/v4"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// ErrHelmNotInstalled indicates that /usr/bin/helm does not exist on the node.
var ErrHelmNotInstalled = errors.New("helm is not installed")

// runs a helm install within the minikube vm or container based on the contents of chart *assets.HelmChart
func installHelmChart(ctx context.Context, chart *assets.HelmChart) *exec.Cmd {
	args := []string{
		fmt.Sprintf("KUBECONFIG=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"helm", "upgrade", "--install", chart.Name, chart.Repo, "--create-namespace",
	}
	if chart.Namespace != "" {
		args = append(args, "--namespace", chart.Namespace)
	}

	if chart.Values != nil {
		for _, value := range chart.Values {
			args = append(args, "--set", value)
		}
	}

	if chart.ValueFiles != nil {
		for _, value := range chart.ValueFiles {
			args = append(args, "--values", value)
		}
	}

	return exec.CommandContext(ctx, "sudo", args...)
}

// runs a helm uninstall based on the contents of chart *assets.HelmChart
func uninstalllHelmChart(ctx context.Context, chart *assets.HelmChart) *exec.Cmd {
	args := []string{
		fmt.Sprintf("KUBECONFIG=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"helm", "uninstall", chart.Name,
	}
	if chart.Namespace != "" {
		args = append(args, "--namespace", chart.Namespace)
	}
	return exec.CommandContext(ctx, "sudo", args...)
}

// based on enable will execute installHelmChart or uninstallHelmChart
func helmUninstallOrInstall(ctx context.Context, chart *assets.HelmChart, enable bool) *exec.Cmd {
	if enable {
		return installHelmChart(ctx, chart)
	}
	return uninstalllHelmChart(ctx, chart)
}

// HelmOptions contains options for installing Helm.
type HelmOptions struct {
	Version *semver.Version
}

// HelmVersion returns the installed helm version at /usr/bin/helm. Returns an
// error if helm is not installed, fails to run, or returns an invalid version
// string.
func HelmVersion(runner command.Runner) (semver.Version, error) {
	rr, err := runner.RunCmd(exec.Command("/usr/bin/helm", "version", "--template", "{{.Version}}"))
	if err != nil {
		if rr.ExitCode == command.NotFound {
			// Expected when starting a new or stopped cluster since /usr/bin/
			// is not persisted.
			stderr := strings.TrimSpace(rr.Stderr.String())
			return semver.Version{}, fmt.Errorf("%w: %s", ErrHelmNotInstalled, stderr)
		}
		return semver.Version{}, fmt.Errorf("failed to run helm version: %w", err)
	}
	raw := strings.TrimPrefix(strings.TrimSpace(rr.Stdout.String()), "v")
	v, err := semver.Parse(raw)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse helm version %q: %w", raw, err)
	}
	return v, nil
}

// InstallHelm installs Helm inside the guest VM/container at /usr/bin/helm.
// Use /usr/bin/helm which is always in the PATH, unlike /usr/local/bin.
func InstallHelm(runner command.Runner, opts HelmOptions) error {
	script := `
		curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
		chmod 700 get_helm.sh
		HELM_INSTALL_DIR=/usr/bin ./get_helm.sh`

	if opts.Version != nil {
		klog.Infof("Installing helm version %s", opts.Version)
		script += " --version v" + opts.Version.String()
	} else {
		klog.Info("Installing helm latest version")
	}

	_, err := runner.RunCmd(exec.Command("sudo", "bash", "-o", "errexit", "-c", script))
	if err != nil {
		return fmt.Errorf("installing helm: %w", err)
	}
	return nil
}
