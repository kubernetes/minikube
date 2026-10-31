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
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/shirou/gopsutil/v4/process"
	"k8s.io/minikube/pkg/minikube/config"
)

// Run the test executable as a small native bridge on every supported OS.
// No cluster, Docker daemon or real provider is used by these process tests.
func TestMain(m *testing.M) {
	if os.Getenv("MINIKUBE_TEST_AUTOSCALER_BRIDGE") == "1" {
		mode, state := "", ""
		for _, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "--mode=") {
				mode = strings.TrimPrefix(arg, "--mode=")
			}
			if strings.HasPrefix(arg, "--state-dir=") {
				state = strings.TrimPrefix(arg, "--state-dir=")
			}
		}
		ready := filepath.Join(state, "ready")
		switch mode {
		case "bridge-check":
			if _, err := os.Stat(ready); err != nil {
				os.Exit(1)
			}
		case "maintenance-check":
			if _, err := os.Stat(ready); err == nil {
				os.Exit(1)
			}
		case "bridge":
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			if err := os.WriteFile(ready, nil, 0600); err != nil {
				os.Exit(1)
			}
			<-ctx.Done()
			_ = os.Remove(ready)
		default:
			os.Exit(2)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestManagedBridgeLifecycle(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	settings := config.ClusterAutoscalerConfig{Binary: binary, StateDir: t.TempDir()}
	t.Setenv("MINIKUBE_TEST_AUTOSCALER_BRIDGE", "1")
	t.Cleanup(func() {
		if err := stopClusterAutoscalerBridge(settings); err != nil {
			t.Error(err)
		}
	})
	if err := startClusterAutoscalerBridge(context.Background(), "test", settings); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(settings.StateDir, "bridge.pid"))
	if err != nil {
		t.Fatal(err)
	}
	if err := startClusterAutoscalerBridge(context.Background(), "test", settings); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(settings.StateDir, "bridge.pid"))
	if err != nil || string(first) != string(second) {
		t.Fatal("re-enable replaced a healthy bridge")
	}
	if err := stopClusterAutoscalerBridge(settings); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(settings.StateDir, "bridge.pid")); !os.IsNotExist(err) {
		t.Fatal("disable retained the PID record")
	}
}

func TestBridgeCleanupDoesNotSignalReusedPID(t *testing.T) {
	settings := config.ClusterAutoscalerConfig{StateDir: t.TempDir()}
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	created, err := p.CreateTime()
	if err != nil {
		t.Fatal(err)
	}
	if err := writeClusterAutoscalerJSON(filepath.Join(settings.StateDir, "bridge.pid"), clusterAutoscalerBridgeProcess{p.Pid, created + 1}); err != nil {
		t.Fatal(err)
	}
	if err := stopClusterAutoscalerBridge(settings); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(settings.StateDir, "bridge.pid")); !os.IsNotExist(err) {
		t.Fatal("stale PID record was not removed")
	}
}
