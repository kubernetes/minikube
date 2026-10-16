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

package autoscaler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gofrs/flock"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/util"
)

const clusterAutoscalerAddon = "cluster-autoscaler"

// Set persists settings and the enabled flag after the lifecycle callback succeeds.
func Set(cc *config.ClusterConfig, name, value string, options *run.CommandOptions) error {
	enable, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	if enable {
		settings, err := clusterAutoscalerSettings(cc, enable)
		if err != nil {
			return err
		}
		cc.ClusterAutoscaler = &settings
	}
	if cc.Addons == nil {
		cc.Addons = make(map[string]bool)
	}
	cc.Addons[name] = enable
	return nil
}

// EnableOrDisable manages the installed provider and the in-cluster autoscaler.
func EnableOrDisable(cc *config.ClusterConfig, _, value string, options *run.CommandOptions) error {
	enable, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("parsing cluster-autoscaler setting: %w", err)
	}
	if !enable && cc.ClusterAutoscaler == nil && !cc.Addons[clusterAutoscalerAddon] {
		return nil
	}
	// Cleanup must remain possible after a Kubernetes upgrade or driver change.
	if enable {
		if err := validateClusterAutoscaler(cc, runtime.GOOS, runtime.GOARCH); err != nil {
			return err
		}
	}
	settings, err := clusterAutoscalerSettings(cc, enable)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(settings.StateDir, 0700); err != nil {
		return err
	}
	lock := flock.New(filepath.Join(settings.StateDir, "minikube.lock"))
	locked, err := lock.TryLock()
	if err != nil || !locked {
		return fmt.Errorf("another cluster-autoscaler lifecycle operation is running (lock error: %v)", err)
	}
	defer lock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	if enable {
		if err := prepareClusterAutoscaler(ctx, cc, settings, options); err != nil {
			return err
		}
		if err := startClusterAutoscalerBridge(ctx, cc.Name, settings); err != nil {
			return err
		}
		if err := runClusterAutoscalerScript(ctx, cc.Name, settings, "enable"); err != nil {
			if !cc.Addons[clusterAutoscalerAddon] {
				// Keep initialized ownership and credentials, but do not leave a
				// new provider making scaling decisions after a failed enable.
				cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				cleanupErr := stopClusterAutoscalerHost(cleanupCtx, cc.Name, settings)
				return fmt.Errorf("enabling cluster-autoscaler: %w", errors.Join(err, cleanupErr))
			}
			return err
		}
		return nil
	}
	if err := runClusterAutoscalerScript(ctx, cc.Name, settings, "disable"); err != nil {
		return err
	}
	return stopClusterAutoscalerBridge(settings)
}

func clusterAutoscalerSettings(cc *config.ClusterConfig, enable bool) (config.ClusterAutoscalerConfig, error) {
	var settings config.ClusterAutoscalerConfig
	if cc.ClusterAutoscaler != nil {
		settings = *cc.ClusterAutoscaler
		// An unrelated shell's environment must not redirect a running addon or
		// its cleanup to a different installation, journal, or provider image.
		if !enable || cc.Addons[clusterAutoscalerAddon] {
			return settings, nil
		}
	}
	for name, target := range map[string]*string{
		"MINIKUBE_AUTOSCALER_ADDON_PATH": &settings.AddonPath,
		"MINIKUBE_AUTOSCALER_STATE_DIR":  &settings.StateDir,
		"MINIKUBE_AUTOSCALER_BINARY":     &settings.Binary,
		"MINIKUBE_AUTOSCALER_IMAGE":      &settings.Image,
	} {
		if value := os.Getenv(name); value != "" {
			*target = value
		}
	}
	if settings.AddonPath == "" {
		settings.AddonPath = discoverClusterAutoscaler()
		if settings.AddonPath == "" {
			return settings, fmt.Errorf("install minikube-cluster-autoscaler-addon in %s or put its bin directory on PATH", localpath.MakeMiniPath("addons", clusterAutoscalerAddon))
		}
	}
	var err error
	settings.AddonPath, err = filepath.Abs(settings.AddonPath)
	if err != nil {
		return settings, err
	}
	if settings.StateDir == "" {
		stateHome := os.Getenv("XDG_STATE_HOME")
		if stateHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return settings, err
			}
			stateHome = filepath.Join(home, ".local", "state")
		}
		settings.StateDir = filepath.Join(stateHome, "minikube-cluster-autoscaler-addon", cc.Name)
	}
	if !filepath.IsAbs(settings.StateDir) {
		return settings, fmt.Errorf("MINIKUBE_AUTOSCALER_STATE_DIR (or XDG_STATE_HOME) must be an absolute path")
	}
	if settings.Binary == "" {
		settings.Binary = filepath.Join(settings.AddonPath, "bin", executableName("minikube-cluster-autoscaler-addon", runtime.GOOS))
	}
	settings.Binary, err = filepath.Abs(settings.Binary)
	if err != nil {
		return settings, err
	}
	if settings.Image == "" {
		settings.Image = "minikube-cluster-autoscaler-addon:local"
	}
	return settings, nil
}

func clusterAutoscalerScriptCommand(ctx context.Context, profile string, settings config.ClusterAutoscalerConfig, action string) *exec.Cmd {
	return lifecycleCommand(ctx, profile, settings, action, runtime.GOOS)
}

func lifecycleCommand(ctx context.Context, profile string, settings config.ClusterAutoscalerConfig, action, goos string) *exec.Cmd {
	var cmd *exec.Cmd
	if goos == "windows" {
		cmd = exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-File", filepath.Join(settings.AddonPath, "scripts", "addon.ps1"), "-Action", action)
	} else {
		cmd = exec.CommandContext(ctx, "bash", filepath.Join(settings.AddonPath, "scripts", "addon.sh"), action)
	}
	cmd.Dir = settings.AddonPath
	cmd.Env = clusterAutoscalerEnvironment(profile, settings)
	return cmd
}

func clusterAutoscalerEnvironment(profile string, settings config.ClusterAutoscalerConfig) []string {
	return append(os.Environ(),
		"MINIKUBE_HOME="+localpath.MiniPath(),
		"PATH="+filepath.Join(settings.StateDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
		"MINIKUBE_AUTOSCALER_PROFILE="+profile,
		"MINIKUBE_AUTOSCALER_STATE_DIR="+settings.StateDir,
		"MINIKUBE_AUTOSCALER_CONFIG="+filepath.Join(settings.StateDir, "config.json"),
		"MINIKUBE_AUTOSCALER_BINARY="+settings.Binary,
		"MINIKUBE_AUTOSCALER_IMAGE="+settings.Image,
	)
}

func runClusterAutoscalerScript(ctx context.Context, profile string, settings config.ClusterAutoscalerConfig, action string) error {
	cmd := clusterAutoscalerScriptCommand(ctx, profile, settings, action)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	_, err := command.NewExecRunner(false).RunCmd(cmd)
	return err
}

func validateClusterAutoscaler(cc *config.ClusterConfig, goos, goarch string) error {
	if (goos != "darwin" && goos != "linux" && goos != "windows") || (goarch != "amd64" && goarch != "arm64") {
		return fmt.Errorf("cluster-autoscaler supports macOS, Linux and Windows on amd64 and arm64; got %s/%s", goos, goarch)
	}
	driver := "kvm2"
	if goos == "darwin" {
		driver = "qemu2"
	} else if goos == "windows" || goarch == "arm64" {
		driver = "docker"
	}
	if cc.Driver != driver {
		return fmt.Errorf("cluster-autoscaler on %s/%s requires the %s driver", goos, goarch, driver)
	}
	if goos == "darwin" && cc.Network != "socket_vmnet" {
		return fmt.Errorf("cluster-autoscaler requires --network=socket_vmnet with qemu2")
	}
	v, err := util.ParseKubernetesVersion(cc.KubernetesConfig.KubernetesVersion)
	if err != nil || v.Major != 1 || v.Minor != 35 {
		return fmt.Errorf("cluster-autoscaler pins Cluster Autoscaler 1.35 and requires Kubernetes v1.35.x; got %q", cc.KubernetesConfig.KubernetesVersion)
	}
	controlPlanes := 0
	for _, n := range cc.Nodes {
		if n.ControlPlane {
			controlPlanes++
		}
	}
	if controlPlanes != 1 {
		return fmt.Errorf("cluster-autoscaler requires exactly one control-plane node; got %d", controlPlanes)
	}
	return nil
}

func executableName(name, goos string) string {
	if goos == "windows" {
		return name + ".exe"
	}
	return name
}
