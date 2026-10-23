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

package autoscaler_test

import (
	"encoding/json"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/run"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func createTestProfile(t *testing.T) string {
	t.Helper()
	td := t.TempDir()

	t.Setenv(localpath.MinikubeHome, td)

	// Not necessary, but it is a handy random alphanumeric
	name := filepath.Base(td)
	if err := os.MkdirAll(config.ProfileFolderPath(name), 0777); err != nil {
		t.Fatalf("error creating temporary directory")
	}

	cc := &config.ClusterConfig{
		Name:             name,
		CPUs:             2,
		Memory:           2500,
		KubernetesConfig: config.KubernetesConfig{},
		Nodes:            []config.Node{{ControlPlane: true}},
	}

	if err := config.DefaultLoader.WriteConfigToFile(name, cc); err != nil {
		t.Fatalf("error creating temporary profile config: %v", err)
	}
	return name
}

// fakeClusterAutoscaler installs a recording script, never the real provider.
func fakeClusterAutoscaler(t *testing.T) (*config.ClusterConfig, string) {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" || runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
		t.Skip("host lifecycle supports macOS/Linux amd64/arm64")
	}
	profile := createTestProfile(t)
	cc, err := config.Load(profile)
	if err != nil {
		t.Fatal(err)
	}
	cc.Driver = "kvm2"
	if runtime.GOOS == "darwin" {
		cc.Driver, cc.Network = "qemu2", "socket_vmnet"
	} else if runtime.GOARCH == "arm64" {
		cc.Driver = "docker"
	}
	cc.KubernetesConfig.KubernetesVersion = "v1.35.0"
	if err := config.Write(profile, cc); err != nil {
		t.Fatal(err)
	}
	// Spaces and shell metacharacters must be treated as literal paths.
	root := filepath.Join(t.TempDir(), "addon $(literal)")
	state := filepath.Join(t.TempDir(), "state directory")
	for _, dir := range []string{filepath.Join(root, "scripts"), filepath.Join(state, "provider"), filepath.Join(root, "tools")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	toolsDir := filepath.Join(root, "tools")
	for _, tool := range []string{"docker", "helm", "kubectl", "jq"} {
		if err := os.WriteFile(filepath.Join(toolsDir, tool), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", toolsDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := os.WriteFile(filepath.Join(root, "custom bridge"), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "provider", "state.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	script := `#!/usr/bin/env bash
set -eu
printf '%s\n' "$1" "$MINIKUBE_AUTOSCALER_PROFILE" "$MINIKUBE_AUTOSCALER_CONFIG" "$MINIKUBE_AUTOSCALER_BINARY" "$MINIKUBE_AUTOSCALER_IMAGE" >> "$MINIKUBE_AUTOSCALER_STATE_DIR/calls"
if [ -f "$MINIKUBE_AUTOSCALER_STATE_DIR/fail" ]; then
    echo 'simulated lifecycle failure' >&2
    exit 17
fi
if [ -n "${MINIKUBE_TEST_PROFILE_REPLACEMENT:-}" ]; then
    cp "$MINIKUBE_TEST_PROFILE_REPLACEMENT" "$MINIKUBE_TEST_PROFILE_CONFIG"
fi
case "$1" in
    build)
        mkdir -p "$(dirname "$MINIKUBE_AUTOSCALER_BINARY")"
        printf '#!/bin/sh\nexit 0\n' > "$MINIKUBE_AUTOSCALER_BINARY"
        chmod 700 "$MINIKUBE_AUTOSCALER_BINARY"
        ;;
    init)
        mkdir -p "$MINIKUBE_AUTOSCALER_STATE_DIR/provider"
        printf '{}' > "$MINIKUBE_AUTOSCALER_STATE_DIR/provider/state.json"
        ;;
esac
`
	if err := os.WriteFile(filepath.Join(root, "scripts", "addon.sh"), []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "config.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MINIKUBE_AUTOSCALER_ADDON_PATH", root)
	t.Setenv("MINIKUBE_AUTOSCALER_STATE_DIR", state)
	t.Setenv("MINIKUBE_AUTOSCALER_BINARY", filepath.Join(root, "custom bridge"))
	t.Setenv("MINIKUBE_AUTOSCALER_IMAGE", "provider:test")
	// These must be overridden by the selected profile and initialized state.
	t.Setenv("MINIKUBE_AUTOSCALER_PROFILE", "wrong-cluster")
	t.Setenv("MINIKUBE_AUTOSCALER_CONFIG", "/wrong/config.json")
	return cc, state
}

func TestClusterAutoscalerLifecycle(t *testing.T) {
	cc, state := fakeClusterAutoscaler(t)
	if assets.Addons["cluster-autoscaler"].IsEnabledOrDefault(cc) {
		t.Fatal("autoscaler must be opt-in")
	}
	if err := addons.SetAndSave(cc.Name, "cluster-autoscaler", "true", &run.CommandOptions{}); err != nil {
		t.Fatal(err)
	}
	cc, err := config.Load(cc.Name)
	if err != nil {
		t.Fatal(err)
	}
	if !cc.Addons["cluster-autoscaler"] || cc.ClusterAutoscaler == nil {
		t.Fatalf("successful enable did not persist settings: %+v", cc)
	}
	saved := *cc.ClusterAutoscaler
	// Restart/disable must use the saved installation even in another shell.
	for _, key := range []string{"ADDON_PATH", "STATE_DIR", "BINARY", "IMAGE"} {
		t.Setenv("MINIKUBE_AUTOSCALER_"+key, "/unrelated/installation")
	}
	if err := addons.SetAndSave(cc.Name, "cluster-autoscaler", "true", &run.CommandOptions{}); err != nil {
		t.Fatal(err)
	}
	// Disabling must work after a version upgrade and must not erase state.
	cc.KubernetesConfig.KubernetesVersion = "v1.37.0"
	if err := config.Write(cc.Name, cc); err != nil {
		t.Fatal(err)
	}
	if err := addons.SetAndSave(cc.Name, "cluster-autoscaler", "false", &run.CommandOptions{}); err != nil {
		t.Fatal(err)
	}
	cc, err = config.Load(cc.Name)
	if err != nil {
		t.Fatal(err)
	}
	if cc.Addons["cluster-autoscaler"] || !reflect.DeepEqual(cc.ClusterAutoscaler, &saved) {
		t.Fatalf("disable should retain settings and clear enabled flag: %+v", cc.ClusterAutoscaler)
	}
	calls, err := os.ReadFile(filepath.Join(state, "calls"))
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, action := range []string{"enable", "enable", "disable"} {
		want = append(want, action, cc.Name, filepath.Join(state, "config.json"), saved.Binary, saved.Image)
	}
	if string(calls) != strings.Join(want, "\n")+"\n" {
		t.Fatalf("lifecycle called with wrong arguments or environment: %s", calls)
	}
}

func TestClusterAutoscalerFailures(t *testing.T) {
	for _, tc := range []string{"enable failure", "disable failure", "incomplete initialization", "unsupported cluster"} {
		t.Run(tc, func(t *testing.T) {
			cc, state := fakeClusterAutoscaler(t)
			value := "true"
			wantError := "simulated lifecycle failure"
			switch tc {
			case "enable failure", "disable failure":
				if tc == "disable failure" {
					if err := addons.SetAndSave(cc.Name, "cluster-autoscaler", "true", &run.CommandOptions{}); err != nil {
						t.Fatal(err)
					}
					value = "false"
				}
				if err := os.WriteFile(filepath.Join(state, "fail"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			case "incomplete initialization":
				if err := os.Remove(filepath.Join(state, "provider", "state.json")); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(state, "host"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(state, "host", "bridge.json"), []byte("{}"), 0600); err != nil {
					t.Fatal(err)
				}
				wantError = "incomplete ownership"
			case "unsupported cluster":
				cc.KubernetesConfig.KubernetesVersion = "v1.37.0"
				if err := config.Write(cc.Name, cc); err != nil {
					t.Fatal(err)
				}
				wantError = "requires Kubernetes v1.35.x"
			}
			err := addons.SetAndSave(cc.Name, "cluster-autoscaler", value, &run.CommandOptions{})
			if err == nil || !strings.Contains(err.Error(), wantError) {
				t.Fatalf("error = %v, want %q", err, wantError)
			}
			cc, err = config.Load(cc.Name)
			if err != nil {
				t.Fatal(err)
			}
			if cc.Addons["cluster-autoscaler"] != (value == "false") {
				t.Fatal("failed lifecycle changed the enabled flag")
			}
			if tc == "incomplete initialization" || tc == "unsupported cluster" {
				if _, err := os.Stat(filepath.Join(state, "calls")); !os.IsNotExist(err) {
					t.Fatal("invalid setup ran the lifecycle script")
				}
			}
		})
	}
}

func TestFirstEnableBuildsAndInitializesOnce(t *testing.T) {
	cc, state := fakeClusterAutoscaler(t)
	if err := os.Remove(os.Getenv("MINIKUBE_AUTOSCALER_BINARY")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(state, "provider", "state.json")); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := addons.SetAndSave(cc.Name, "cluster-autoscaler", "true", &run.CommandOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	calls, err := os.ReadFile(filepath.Join(state, "calls"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(calls)), "\n")
	var actions []string
	for i := 0; i < len(lines); i += 5 {
		actions = append(actions, lines[i])
	}
	if !reflect.DeepEqual(actions, []string{"build", "init", "enable", "enable"}) {
		t.Fatalf("lifecycle actions = %v", actions)
	}
}

func TestEnablePreservesNewWorkerInventory(t *testing.T) {
	cc, state := fakeClusterAutoscaler(t)
	cc.Nodes = append(cc.Nodes, config.Node{Name: "m02", Worker: true})
	data, err := json.Marshal(cc)
	if err != nil {
		t.Fatal(err)
	}
	replacement := filepath.Join(state, "updated-profile.json")
	if err := os.WriteFile(replacement, data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MINIKUBE_TEST_PROFILE_REPLACEMENT", replacement)
	t.Setenv("MINIKUBE_TEST_PROFILE_CONFIG", filepath.Join(config.ProfileFolderPath(cc.Name), "config.json"))
	if err := addons.SetAndSave(cc.Name, "cluster-autoscaler", "true", &run.CommandOptions{}); err != nil {
		t.Fatal(err)
	}
	got, err := config.Load(cc.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Nodes) != 2 || got.Nodes[1].Name != "m02" || !got.Addons["cluster-autoscaler"] {
		t.Fatalf("enable overwrote provider updates: %+v", got)
	}
}
