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
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/run"
)

func discoverClusterAutoscaler() string {
	candidates := []string{localpath.MakeMiniPath("addons", clusterAutoscalerAddon)}
	if binary, err := exec.LookPath("minikube-cluster-autoscaler-addon"); err == nil {
		if resolved, err := filepath.EvalSymlinks(binary); err == nil {
			candidates = append(candidates, filepath.Dir(filepath.Dir(resolved)))
		}
	}
	// Support developing the two projects side by side, including out/minikube.
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "..", "minikube-cluster-autoscaler-addon"))
	}
	if binary, err := os.Executable(); err == nil {
		for _, parent := range []string{"..", "../.."} {
			candidates = append(candidates, filepath.Join(filepath.Dir(binary), parent, "minikube-cluster-autoscaler-addon"))
		}
	}
	for _, root := range candidates {
		script := "addon.sh"
		if runtime.GOOS == "windows" {
			script = "addon.ps1"
		}
		if info, err := os.Stat(filepath.Join(root, "scripts", script)); err == nil && info.Mode().IsRegular() {
			return root
		}
	}
	return ""
}

// These are the installed project's JSON settings, not a Go dependency on it.
type clusterAutoscalerDefaults struct {
	Profile                 string `json:"profile"`
	Namespace               string `json:"namespace"`
	Listen                  string `json:"listen"`
	MinWorkers              int    `json:"minWorkers"`
	MaxWorkers              int    `json:"maxWorkers"`
	MaxTotalMemoryMiB       int    `json:"maxTotalMemoryMiB"`
	ProvisionTimeoutSeconds int    `json:"provisionTimeoutSeconds"`
	CooldownSeconds         int    `json:"cooldownSeconds"`
}

func defaultClusterAutoscalerConfig(cc *config.ClusterConfig, hostIP net.IP, hostMemoryMiB uint64) (clusterAutoscalerDefaults, error) {
	c := clusterAutoscalerDefaults{
		Profile: cc.Name, Namespace: "kube-system", MinWorkers: 0,
		ProvisionTimeoutSeconds: 900, CooldownSeconds: 30,
	}
	if hostIP == nil || hostIP.IsUnspecified() || hostIP.IsLoopback() || hostIP.IsMulticast() {
		return c, fmt.Errorf("cannot determine a reachable host IP for cluster-autoscaler")
	}
	c.Listen = net.JoinHostPort(hostIP.String(), "50051")
	if cc.Memory <= 0 || len(cc.Nodes) == 0 {
		return c, fmt.Errorf("cluster-autoscaler requires a running cluster with known node memory")
	}
	// Reserve at least a quarter of physical memory for the host and provider.
	budget := int(hostMemoryMiB * 3 / 4)
	c.MaxWorkers = min(2, budget/cc.Memory-len(cc.Nodes))
	if c.MaxWorkers < 1 {
		return c, fmt.Errorf("not enough host memory for an additional worker; use a cluster with smaller --memory")
	}
	c.MaxTotalMemoryMiB = max(4096, (len(cc.Nodes)+c.MaxWorkers)*cc.Memory)
	return c, nil
}

func prepareClusterAutoscaler(ctx context.Context, cc *config.ClusterConfig, settings config.ClusterAutoscalerConfig, options *run.CommandOptions) error {
	dependencies := []string{"docker", "helm", "kubectl"}
	if runtime.GOOS == "windows" {
		dependencies = append(dependencies, "powershell.exe")
	} else {
		dependencies = append(dependencies, "bash", "jq")
	}
	for _, binary := range dependencies {
		if _, err := exec.LookPath(binary); err != nil {
			return fmt.Errorf("cluster-autoscaler requires %s on PATH: %w", binary, err)
		}
	}
	// The bridge's node operations must use this minikube binary and home.
	binDir := filepath.Join(settings.StateDir, "bin")
	if err := os.MkdirAll(binDir, 0700); err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	if err := installMinikubeBinary(binary, filepath.Join(binDir, executableName("minikube", runtime.GOOS))); err != nil {
		return err
	}
	if _, err := os.Stat(settings.Binary); err != nil || exec.CommandContext(ctx, "docker", "image", "inspect", settings.Image).Run() != nil {
		out.Infof("Preparing Cluster Autoscaler host components...")
		if err := runClusterAutoscalerScript(ctx, cc.Name, settings, "build"); err != nil {
			return err
		}
	}
	configPath := filepath.Join(settings.StateDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		for _, journal := range []string{"provider/state.json", "host/bridge.json"} {
			if _, err := os.Stat(filepath.Join(settings.StateDir, journal)); !os.IsNotExist(err) {
				return fmt.Errorf("cluster-autoscaler ownership journals exist without config.json; restore the original configuration")
			}
		}
		api, err := machine.NewAPIClient(options)
		if err != nil {
			return err
		}
		defer api.Close()
		cp, err := config.ControlPlane(*cc)
		if err != nil {
			return err
		}
		host, err := machine.LoadHost(api, config.MachineName(*cc, cp))
		if err != nil {
			return err
		}
		hostIP, err := cluster.HostIP(host, cc.Name)
		if err != nil {
			return err
		}
		memory, err := mem.VirtualMemory()
		if err != nil {
			return err
		}
		memoryBytes := memory.Total
		if cc.Driver == "docker" {
			daemon, err := oci.CachedDaemonInfo(oci.Docker)
			if err != nil {
				return err
			}
			if daemon.OSType != "linux" || daemon.TotalMemory <= 0 {
				return fmt.Errorf("cluster-autoscaler requires Docker with Linux containers and available node memory")
			}
			memoryBytes = min(memoryBytes, uint64(daemon.TotalMemory))
		}
		defaults, err := defaultClusterAutoscalerConfig(cc, hostIP, memoryBytes/(1024*1024))
		if err != nil {
			return err
		}
		if err := writeClusterAutoscalerJSON(configPath, defaults); err != nil {
			return err
		}
		out.Infof("Cluster Autoscaler will allow up to {{.workers}} additional workers.", out.V{"workers": defaults.MaxWorkers})
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(settings.StateDir, "provider", "state.json")); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(settings.StateDir, "host", "bridge.json")); !os.IsNotExist(err) {
			return fmt.Errorf("cluster-autoscaler has incomplete ownership state; restore its journals before enabling")
		}
		out.Infof("Initializing Cluster Autoscaler...")
		return runClusterAutoscalerScript(ctx, cc.Name, settings, "init")
	} else {
		return err
	}
}

// Use a hard link where possible. Windows symlinks require privileges that an
// addon should not need; copying also handles installations on another volume.
func installMinikubeBinary(source, destination string) error {
	srcInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	if dstInfo, err := os.Stat(destination); err == nil && os.SameFile(srcInfo, dstInfo) {
		return nil
	}
	if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Link(source, destination); err == nil {
		return nil
	}
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(destination)
		return copyErr
	}
	return closeErr
}

func writeClusterAutoscalerJSON(path string, value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}

func clusterAutoscalerBridgeCommand(ctx context.Context, profile string, settings config.ClusterAutoscalerConfig, mode string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, settings.Binary, "--mode="+mode,
		"--config="+filepath.Join(settings.StateDir, "config.json"), "--state-dir="+settings.StateDir)
	cmd.Env = clusterAutoscalerEnvironment(profile, settings)
	return cmd
}

type clusterAutoscalerBridgeProcess struct {
	PID     int32
	Created int64
}

func startClusterAutoscalerBridge(ctx context.Context, profile string, settings config.ClusterAutoscalerConfig) error {
	if clusterAutoscalerBridgeCommand(ctx, profile, settings, "bridge-check").Run() == nil {
		return nil
	}
	// A running but unhealthy bridge must not be replaced or have its PID lost.
	if output, err := clusterAutoscalerBridgeCommand(ctx, profile, settings, "maintenance-check").CombinedOutput(); err != nil {
		return fmt.Errorf("cluster-autoscaler bridge is unavailable: %s: %w", output, err)
	}
	logPath := filepath.Join(settings.StateDir, "bridge.log")
	log, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer log.Close()
	// The bridge deliberately outlives this command; readiness is bounded below.
	cmd := clusterAutoscalerBridgeCommand(context.Background(), profile, settings, "bridge")
	detachClusterAutoscalerBridge(cmd)
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait() // reap the process if it exits while minikube is still running
	p, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("starting cluster-autoscaler bridge; see %s: %w", logPath, err)
	}
	created, err := p.CreateTime()
	if err != nil {
		_ = cmd.Process.Kill()
		return err
	}
	if err := writeClusterAutoscalerJSON(filepath.Join(settings.StateDir, "bridge.pid"), clusterAutoscalerBridgeProcess{p.Pid, created}); err != nil {
		_ = cmd.Process.Kill()
		return err
	}
	checkCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	for checkCtx.Err() == nil {
		if clusterAutoscalerBridgeCommand(checkCtx, profile, settings, "bridge-check").Run() == nil {
			return nil
		}
		if running, _ := p.IsRunning(); !running {
			break
		}
		select {
		case <-checkCtx.Done():
		case <-time.After(200 * time.Millisecond):
		}
	}
	_ = stopClusterAutoscalerBridge(settings)
	return fmt.Errorf("cluster-autoscaler bridge did not become ready; see %s", logPath)
}

func stopClusterAutoscalerBridge(settings config.ClusterAutoscalerConfig) error {
	path := filepath.Join(settings.StateDir, "bridge.pid")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil // a manually started bridge is owned by its caller
	}
	if err != nil {
		return err
	}
	var record clusterAutoscalerBridgeProcess
	if err := json.Unmarshal(data, &record); err != nil || record.PID <= 0 || record.Created <= 0 {
		return fmt.Errorf("invalid cluster-autoscaler bridge PID record: %s", path)
	}
	p, err := process.NewProcess(record.PID)
	if err == process.ErrorProcessNotRunning {
		return os.Remove(path)
	}
	if err != nil {
		return err
	}
	created, err := p.CreateTime()
	if err != nil {
		return err
	}
	if created != record.Created {
		return os.Remove(path) // the PID has been reused; never signal it
	}
	if err := p.Terminate(); err != nil {
		return err
	}
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
		if running, _ := p.IsRunning(); !running {
			return os.Remove(path)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cluster-autoscaler bridge did not stop; see %s", filepath.Join(settings.StateDir, "bridge.log"))
}

// Stop stops host processes before cluster stop/delete while
// retaining the addon flag so the next start can re-enable it automatically.
func Stop(cc *config.ClusterConfig) error {
	if cc == nil || cc.ClusterAutoscaler == nil {
		return nil
	}
	if !cc.Addons[clusterAutoscalerAddon] {
		return stopClusterAutoscalerBridge(*cc.ClusterAutoscaler)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return stopClusterAutoscalerHost(ctx, cc.Name, *cc.ClusterAutoscaler)
}

func stopClusterAutoscalerHost(ctx context.Context, profile string, settings config.ClusterAutoscalerConfig) error {
	container := "minikube-cluster-autoscaler-addon-" + profile
	cmd := exec.CommandContext(ctx, "docker", "container", "ls", "--all", "--filter", "name=^/"+container+"$", "--format", "{{.ID}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("finding cluster-autoscaler provider: %s: %w", output, err)
	}
	if strings.TrimSpace(string(output)) != "" {
		if output, err := exec.CommandContext(ctx, "docker", "stop", "--time", "30", container).CombinedOutput(); err != nil {
			return fmt.Errorf("stopping cluster-autoscaler provider: %s: %w", output, err)
		}
	}
	return stopClusterAutoscalerBridge(settings)
}
