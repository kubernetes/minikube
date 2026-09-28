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
	"k8s.io/minikube/pkg/minikube/config"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestValidateClusterAutoscaler(t *testing.T) {
	for _, tc := range []struct {
		name, goos, goarch, driver, network, version string
		controlPlanes                                int
		wantError                                    string
	}{
		{"macOS arm64", "darwin", "arm64", "qemu2", "socket_vmnet", "v1.35.0", 1, ""},
		{"macOS amd64", "darwin", "amd64", "qemu2", "socket_vmnet", "v1.35.4", 1, ""},
		{"Linux amd64", "linux", "amd64", "kvm2", "", "v1.35.0", 1, ""},
		{"Linux arm64", "linux", "arm64", "docker", "", "v1.35.0", 1, ""},
		{"Windows amd64", "windows", "amd64", "docker", "", "v1.35.0", 1, ""},
		{"Windows arm64", "windows", "arm64", "docker", "", "v1.35.0", 1, ""},
		{"Windows Hyper-V", "windows", "amd64", "hyperv", "", "v1.35.0", 1, "docker driver"},
		{"unsupported architecture", "linux", "ppc64le", "docker", "", "v1.35.0", 1, "supports macOS, Linux and Windows"},
		{"macOS driver", "darwin", "arm64", "docker", "", "v1.35.0", 1, "qemu2 driver"},
		{"Linux amd64 driver", "linux", "amd64", "docker", "", "v1.35.0", 1, "kvm2 driver"},
		{"Linux arm64 driver", "linux", "arm64", "kvm2", "", "v1.35.0", 1, "docker driver"},
		{"macOS network", "darwin", "arm64", "qemu2", "user", "v1.35.0", 1, "socket_vmnet"},
		{"older Kubernetes", "linux", "amd64", "kvm2", "", "v1.34.0", 1, "requires Kubernetes v1.35.x"},
		{"newer Kubernetes", "linux", "amd64", "kvm2", "", "v1.37.0", 1, "requires Kubernetes v1.35.x"},
		{"invalid version", "linux", "amd64", "kvm2", "", "invalid", 1, "requires Kubernetes v1.35.x"},
		{"missing control plane", "linux", "amd64", "kvm2", "", "v1.35.0", 0, "exactly one control-plane"},
		{"HA", "linux", "amd64", "kvm2", "", "v1.35.0", 2, "exactly one control-plane"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cc := &config.ClusterConfig{
				Driver: tc.driver, Network: tc.network,
				KubernetesConfig: config.KubernetesConfig{KubernetesVersion: tc.version},
				Nodes:            []config.Node{{Worker: true}},
			}
			for range tc.controlPlanes {
				cc.Nodes = append(cc.Nodes, config.Node{ControlPlane: true})
			}
			err := validateClusterAutoscaler(cc, tc.goos, tc.goarch)
			if tc.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error = %v, want %q", err, tc.wantError)
			}
		})
	}
}

func TestClusterAutoscalerSettings(t *testing.T) {
	for _, key := range []string{"ADDON_PATH", "STATE_DIR", "BINARY", "IMAGE"} {
		t.Setenv("MINIKUBE_AUTOSCALER_"+key, "")
	}
	t.Chdir(t.TempDir())
	cc := &config.ClusterConfig{Name: "custom-profile"}
	if _, err := clusterAutoscalerSettings(cc, true); err == nil {
		t.Fatal("missing installation must fail")
	}
	root := t.TempDir()
	t.Setenv("MINIKUBE_AUTOSCALER_ADDON_PATH", root)
	t.Setenv("XDG_STATE_HOME", root)
	settings, err := clusterAutoscalerSettings(cc, true)
	if err != nil {
		t.Fatal(err)
	}
	want := config.ClusterAutoscalerConfig{
		AddonPath: root,
		StateDir:  filepath.Join(root, "minikube-cluster-autoscaler-addon", cc.Name),
		Binary:    filepath.Join(root, "bin", executableName("minikube-cluster-autoscaler-addon", runtime.GOOS)),
		Image:     "minikube-cluster-autoscaler-addon:local",
	}
	if settings != want {
		t.Fatalf("settings = %+v, want %+v", settings, want)
	}
	t.Setenv("MINIKUBE_AUTOSCALER_STATE_DIR", "relative/path")
	if _, err := clusterAutoscalerSettings(cc, true); err == nil {
		t.Fatal("relative state directory must fail")
	}
}

func TestDefaultConfiguration(t *testing.T) {
	cc := &config.ClusterConfig{Name: "demo", Memory: 4096, Nodes: []config.Node{{ControlPlane: true}}}
	for _, tc := range []struct {
		memory  uint64
		workers int
	}{{16384, 2}, {12288, 1}, {8192, 0}} {
		cfg, err := defaultClusterAutoscalerConfig(cc, net.ParseIP("192.168.105.1"), tc.memory)
		if tc.workers == 0 {
			if err == nil {
				t.Fatal("accepted insufficient host memory")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Profile != cc.Name || cfg.MaxWorkers != tc.workers || cfg.MinWorkers != 0 || cfg.MaxTotalMemoryMiB != (1+tc.workers)*4096 || cfg.Listen != "192.168.105.1:50051" {
			t.Fatalf("unexpected defaults: %+v", cfg)
		}
	}
	if _, err := defaultClusterAutoscalerConfig(cc, nil, 16384); err == nil {
		t.Fatal("accepted unknown host address")
	}
}

func TestWindowsLifecycleCommand(t *testing.T) {
	settings := config.ClusterAutoscalerConfig{AddonPath: filepath.Join(t.TempDir(), "addon with spaces"), StateDir: t.TempDir()}
	cmd := lifecycleCommand(context.Background(), "selected-profile", settings, "enable", "windows")
	want := []string{"powershell.exe", "-NoProfile", "-NonInteractive", "-File", filepath.Join(settings.AddonPath, "scripts", "addon.ps1"), "-Action", "enable"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("command = %v, want %v", cmd.Args, want)
	}
	if !strings.Contains(strings.Join(cmd.Env, "\n"), "MINIKUBE_AUTOSCALER_PROFILE=selected-profile") {
		t.Fatal("selected profile not passed to Windows lifecycle")
	}
	if executableName("provider", "windows") != "provider.exe" {
		t.Fatal("Windows executable suffix missing")
	}
}

func TestInstallMinikubeBinary(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source")
	dst := filepath.Join(t.TempDir(), "minikube.exe")
	if err := os.WriteFile(src, []byte("test binary"), 0700); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := installMinikubeBinary(src, dst); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "test binary" {
		t.Fatalf("installed binary = %q, %v", data, err)
	}
}
