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
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// runs a helm install/uninstall based on the contents of chart *assets.HelmChart and enable
func helmCommand(ctx context.Context, chart *assets.HelmChart, enable bool) *exec.Cmd {
	var args []string

	if !enable {
		args = []string{
			fmt.Sprintf("KUBECONFIG=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
			"helm", "uninstall", chart.Name,
		}
		if chart.Namespace != "" {
			args = append(args, "--namespace", chart.Namespace)
		}
		return exec.CommandContext(ctx, "sudo", args...)
	}

	args = []string{
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

func helmInstallBinary(addon *assets.Addon, runner command.Runner) error {
	_, err := runner.RunCmd(exec.Command("test", "-f", "/usr/bin/helm"))
	if err != nil {
		// If not, install it
		rr, err := runner.RunCmd(exec.Command("uname", "-m"))
		if err != nil {
			return errors.Wrap(err, "getting architecture")
		}
	arch := strings.TrimSpace(rr.Stdout.String())
	var helmArch string
	switch arch {
		case "x86_64":
			helmArch = "amd64"
		case "aarch64", "arm64":
			helmArch = "arm64"
		default:
			return fmt.Errorf("failure to detect architecture or unsupported architecture: %s", arch)
	}
	helmURL := fmt.Sprintf("https://get.helm.sh/helm-v3.19.0-linux-%s.tar.gz", helmArch)
	installCmd := fmt.Sprintf("curl -sSL %s | tar -xzf - -C /usr/bin --strip-components=1 linux-%s/helm", helmURL, helmArch)
	_, err = runner.RunCmd(exec.Command("sudo", "bash", "-c", installCmd))
	if err != nil {
		return errors.Wrap(err, "installing helm")
	}
	return err
	}
return err
}