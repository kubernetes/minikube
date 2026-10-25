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

package cmd

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

func nodeFlagCommand(t *testing.T, args ...string) (*cobra.Command, *viper.Viper) {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().StringArray(nodeSpec, nil, "")
	cmd.Flags().IntP(nodes, "n", 1, "")
	cmd.Flags().Bool(ha, false, "")
	cmd.Flags().String("driver", "", "")
	cmd.Flags().String("vm-driver", "", "")
	cmd.Flags().String(cniFlag, "", "")
	cmd.Flags().String(containerRuntime, "", "")
	cmd.Flags().Bool(noKubernetes, false, "")
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	v := viper.New()
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		t.Fatal(err)
	}
	return cmd, v
}

func TestNodeFlagRegistration(t *testing.T) {
	f := startCmd.Flags().Lookup(nodeSpec)
	if f == nil || f.Value.Type() != "stringArray" || f.Hidden {
		t.Fatalf("--node must be a visible, repeatable stringArray flag: %+v", f)
	}
	if startCmd.Flags().Lookup("node-os") != nil {
		t.Fatal("obsolete --node-os flag is still registered")
	}
}

func TestNodeFlagResolution(t *testing.T) {
	cp := config.Node{ControlPlane: true, Worker: true}
	worker := config.Node{Worker: true}
	win := config.Node{Worker: true, Guest: config.Guest{Name: config.GuestOSWindows}}
	for _, tc := range []struct {
		name string
		args []string
		want []config.Node
		err  string
	}{
		{name: "explicit linux", args: []string{"--node", "role=control-plane", "--node", "role=worker"}, want: []config.Node{cp, worker}},
		{name: "mixed", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows"}, want: []config.Node{cp, win}},
		{name: "worker first", args: []string{"--node=os=windows,role=worker", "--node=role=control-plane"}, want: []config.Node{cp, win}},
		{name: "case and space", args: []string{"--node= Role = Control-Plane , OS=Linux ", "--node=role=Worker,os=WINDOWS"}, want: []config.Node{cp, win}},
		{name: "parser supports future workers", args: []string{"--node=role=worker,os=windows", "--node=role=control-plane", "--node=role=worker", "--node=role=worker,os=windows"}, want: []config.Node{cp, win, worker, win}},
		{name: "count conflict", args: []string{"--node=role=control-plane", "-n", "1"}, err: "cannot be combined"},
		{name: "HA conflict", args: []string{"--node=role=control-plane", "--ha"}, err: "cannot be combined"},
		{name: "explicit false HA conflict", args: []string{"--node=role=control-plane", "--ha=false"}, err: "cannot be combined"},
		{name: "empty", args: []string{"--node="}, err: "at least one node specification"},
		{name: "empty second node", args: []string{"--node=role=control-plane", "--node="}, err: "nonempty key=value"},
		{name: "missing role", args: []string{"--node=os=linux"}, err: "role is required"},
		{name: "empty value", args: []string{"--node=role="}, err: "nonempty key=value"},
		{name: "missing equals", args: []string{"--node=control-plane"}, err: "nonempty key=value"},
		{name: "duplicate role", args: []string{"--node=role=control-plane,role=worker"}, err: `duplicate key "role"`},
		{name: "duplicate OS case", args: []string{"--node=role=worker,os=linux,OS=windows"}, err: `duplicate key "os"`},
		{name: "unknown key", args: []string{"--node=role=control-plane,name=foo"}, err: `unknown key "name"`},
		{name: "unknown role", args: []string{"--node=role=master"}, err: "unsupported role"},
		{name: "unknown OS", args: []string{"--node=role=worker,os=darwin"}, err: "unsupported OS"},
		{name: "trailing comma", args: []string{"--node=role=control-plane,"}, err: "nonempty key=value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, v := nodeFlagCommand(t, tc.args...)
			before := v.AllSettings()
			got, err := resolveNodes(cmd, nil, v)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v; want %q", err, tc.err)
				}
			} else if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("nodes = %+v, error = %v; want %+v", got, err, tc.want)
			}
			if !reflect.DeepEqual(before, v.AllSettings()) {
				t.Fatal("resolving nodes changed Viper settings")
			}
		})
	}
}

func TestNodeDefaults(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		hostOS string
		err    string
	}{
		{name: "linux independent of host", args: []string{"--node=role=control-plane", "--node=role=worker", "--driver=docker"}, hostOS: "darwin"},
		{name: "mixed defaults", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows"}, hostOS: "windows"},
		{name: "mixed explicit defaults", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--driver=hyperv", "--cni=flannel", "--container-runtime=containerd"}, hostOS: "windows"},
		{name: "Windows host guard", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows"}, hostOS: "linux", err: "Windows host"},
		{name: "Windows control plane", args: []string{"--node=role=control-plane,os=windows"}, hostOS: "windows", err: "Windows control-plane"},
		{name: "missing control plane", args: []string{"--node=role=worker"}, hostOS: "windows", err: "requires a Linux control-plane"},
		{name: "two control planes", args: []string{"--node=role=control-plane", "--node=role=control-plane"}, hostOS: "windows", err: "currently supports one Linux control-plane"},
		{name: "HA via node deferred", args: []string{"--node=role=control-plane", "--node=role=worker", "--node=role=control-plane", "--node=role=control-plane"}, hostOS: "linux", err: "currently supports one Linux control-plane"},
		{name: "too many Windows workers", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--node=role=worker,os=windows"}, hostOS: "windows", err: "currently support"},
		{name: "extra Linux worker", args: []string{"--node=role=control-plane", "--node=role=worker", "--node=role=worker,os=windows"}, hostOS: "windows", err: "currently support"},
		{name: "driver conflict", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--driver=docker"}, hostOS: "windows", err: "--driver=hyperv"},
		{name: "deprecated driver conflict", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--vm-driver=docker"}, hostOS: "windows", err: "--vm-driver=docker"},
		{name: "CNI conflict after driver check", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--cni=calico"}, hostOS: "windows", err: "--cni=flannel"},
		{name: "runtime conflict after CNI check", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--container-runtime=docker"}, hostOS: "windows", err: "--container-runtime=containerd"},
		{name: "no Kubernetes", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--no-kubernetes"}, hostOS: "windows", err: "--no-kubernetes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, v := nodeFlagCommand(t, tc.args...)
			ns, err := resolveNodes(cmd, nil, v)
			if err != nil {
				t.Fatal(err)
			}
			before := v.AllSettings()
			err = applyNodeDefaults(cmd, nil, ns, tc.hostOS, v)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v; want %q", err, tc.err)
				}
				if !reflect.DeepEqual(before, v.AllSettings()) {
					t.Fatal("invalid topology or flag conflict mutated global settings")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if v.GetInt(nodes) != len(ns) {
				t.Fatal("resource sizing will not see the resolved node count")
			}
			if v.Get(ha) != before[ha] {
				t.Fatal("--node changed the HA setting")
			}
			if config.HasWindowsNodes(ns) {
				if v.GetString("driver") != "hyperv" || v.GetString(cniFlag) != "flannel" || v.GetString(containerRuntime) != "containerd" {
					t.Fatal("Windows defaults not applied")
				}
			} else {
				for _, key := range []string{"driver", cniFlag, containerRuntime} {
					if v.Get(key) != before[key] {
						t.Fatalf("Linux topology changed %q", key)
					}
				}
			}
		})
	}
}

func TestNodeSavedTopology(t *testing.T) {
	saved := config.ClusterConfig{
		Name: "saved", Driver: "hyperv",
		KubernetesConfig: config.KubernetesConfig{CNI: "flannel", ContainerRuntime: "containerd"},
		Nodes: []config.Node{
			{Name: "", ControlPlane: true, Worker: true, IP: "192.0.2.1"},
			{Name: "m02", Worker: true, IP: "192.0.2.2", Guest: config.Guest{Name: "windows", Version: "2025", URL: "file:///C:/image.vhdx"}},
		},
	}
	for _, tc := range []struct {
		name string
		args []string
		err  string
	}{
		{name: "repeat same topology", args: []string{"--node=role=worker,os=windows", "--node=role=control-plane"}},
		{name: "different count", args: []string{"--node=role=control-plane"}, err: "cannot change the topology"},
		{name: "different OS", args: []string{"--node=role=control-plane", "--node=role=worker"}, err: "cannot change the topology"},
		{name: "CNI mutation", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--cni=calico"}, err: "--cni=flannel"},
		{name: "driver mutation", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--driver=docker"}, err: "--driver=hyperv"},
		{name: "runtime mutation", args: []string{"--node=role=control-plane", "--node=role=worker,os=windows", "--container-runtime=docker"}, err: "--container-runtime=containerd"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, v := nodeFlagCommand(t, tc.args...)
			before, err := json.Marshal(saved)
			if err != nil {
				t.Fatal(err)
			}
			ns, err := resolveNodes(cmd, &saved, v)
			if err == nil {
				err = applyNodeDefaults(cmd, &saved, ns, "windows", v)
			}
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v; want %q", err, tc.err)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(ns, saved.Nodes) {
				t.Fatalf("saved topology changed: %+v, %v", ns, err)
			}
			ns[1].Guest.URL = "changed"
			after, _ := json.Marshal(saved)
			if string(before) != string(after) {
				t.Fatal("resolved nodes alias the saved profile")
			}
		})
	}
}

func TestNodeLegacyInputsUnchanged(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "default"},
		{name: "Linux multi-node", args: []string{"-n", "2"}},
		{name: "HA", args: []string{"--ha"}},
		{name: "HA with workers", args: []string{"--ha", "-n", "5"}},
		{name: "invalid HA count left to existing validation", args: []string{"--ha", "-n", "2"}},
		{name: "zero count left to existing validation", args: []string{"-n", "0"}},
		{name: "negative count left to existing validation", args: []string{"-n", "-1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, v := nodeFlagCommand(t, tc.args...)
			before := v.AllSettings()
			for _, existing := range []*config.ClusterConfig{
				nil,
				{Name: "saved-linux", Nodes: []config.Node{{ControlPlane: true}, {Name: "m02"}}},
				{Name: "saved-ha", Nodes: []config.Node{{ControlPlane: true}, {Name: "m02", ControlPlane: true}, {Name: "m03", ControlPlane: true}}},
				{Name: "saved-mixed", Nodes: []config.Node{{ControlPlane: true}, {Name: "m02", Guest: config.Guest{Name: "windows"}}}},
			} {
				ns, err := resolveNodes(cmd, existing, v)
				if err != nil || ns != nil {
					t.Fatalf("legacy input intercepted: %+v, %v", ns, err)
				}
				if err := applyNodeDefaults(cmd, existing, ns, "linux", v); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(before, v.AllSettings()) {
					t.Fatal("--node handling changed existing flag/config behavior")
				}
			}
		})
	}
}

func TestNodeExplicitTopologyOverridesDefaults(t *testing.T) {
	cmd, v := nodeFlagCommand(t, "--node=role=control-plane", "--node=role=worker")
	v.SetDefault(nodes, 5)
	ns, err := resolveNodes(cmd, nil, v)
	if err != nil {
		t.Fatal(err)
	}
	if err := applyNodeDefaults(cmd, nil, ns, "linux", v); err != nil {
		t.Fatal(err)
	}
	if len(ns) != 2 || v.GetInt(nodes) != 2 || v.GetBool(ha) {
		t.Fatalf("implicit defaults overrode explicit nodes: %+v", ns)
	}
}

func TestNodeRejectsImplicitHA(t *testing.T) {
	cmd, v := nodeFlagCommand(t, "--node=role=control-plane", "--node=role=worker")
	v.SetDefault(ha, true)
	before := v.AllSettings()
	if _, err := resolveNodes(cmd, nil, v); err == nil || !strings.Contains(err.Error(), "HA to be disabled") {
		t.Fatalf("expected an implicit HA conflict, got %v", err)
	}
	if !reflect.DeepEqual(before, v.AllSettings()) {
		t.Fatal("--node silently changed the configured HA setting")
	}
}

func TestNodeLegacyProfileJSON(t *testing.T) {
	var saved config.ClusterConfig
	raw := `{"Name":"legacy","NodeOS":["linux","windows"],"Driver":"hyperv","KubernetesConfig":{"CNI":"flannel","ContainerRuntime":"containerd"},"Nodes":[{"ControlPlane":true,"Worker":true},{"Name":"m02","Worker":true,"Guest":{"Name":"windows","Version":"2025","URL":"saved.vhdx"}}]}`
	if err := json.Unmarshal([]byte(raw), &saved); err != nil {
		t.Fatal(err)
	}
	cmd, v := nodeFlagCommand(t, "--node=role=control-plane", "--node=role=worker,os=windows")
	ns, err := resolveNodes(cmd, &saved, v)
	if err != nil {
		t.Fatal(err)
	}
	if err := applyNodeDefaults(cmd, &saved, ns, "windows", v); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ns, saved.Nodes) || !config.HasWindowsNodes(ns) {
		t.Fatal("legacy profile lost its Windows node definition")
	}
}

func TestNodeConfiguredTopology(t *testing.T) {
	for key, value := range map[string]any{kubernetesVersion: "", containerRuntime: "", noKubernetes: false, windowsVhdURL: `C:\images\worker.vhdx`} {
		old := viper.Get(key)
		viper.Set(key, value)
		t.Cleanup(func() { viper.Set(key, old) })
	}
	cmd, v := nodeFlagCommand(t, "--node=role=worker,os=windows", "--node=role=control-plane")
	ns, err := resolveNodes(cmd, nil, v)
	if err != nil {
		t.Fatal(err)
	}
	cc := config.ClusterConfig{
		Name: "topology", APIServerPort: 8443, Nodes: slices.Clone(ns),
		KubernetesConfig: config.KubernetesConfig{KubernetesVersion: constants.DefaultKubernetesVersion, ContainerRuntime: "containerd"},
	}
	cc, primary, err := configureNodes(cc, nil)
	if err != nil {
		t.Fatal(err)
	}
	cc.Nodes = configureNodeSpecs(ns, primary)
	if len(cc.Nodes) != 2 || !config.IsPrimaryControlPlane(cc, primary) || primary.Name != "" || cc.Nodes[1].Name != "m02" {
		t.Fatalf("wrong topology or naming: %+v", cc.Nodes)
	}
	if !config.HasWindowsNodes(cc.Nodes) || cc.Nodes[1].Guest.Version != "2025" || cc.Nodes[1].Guest.URL != `C:\images\worker.vhdx` {
		t.Fatalf("Windows image configuration lost: %+v", cc.Nodes[1])
	}
	for _, n := range cc.Nodes {
		if n.Port != 8443 || n.ContainerRuntime != "containerd" || n.KubernetesVersion != constants.DefaultKubernetesVersion {
			t.Fatalf("node not fully configured: %+v", n)
		}
	}
	raw, err := json.Marshal(cc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "NodeOS") {
		t.Fatal("positional OS list is still persisted")
	}
	var restored config.ClusterConfig
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	restart, err := resolveNodes(cmd, &restored, v)
	if err != nil || !reflect.DeepEqual(restart, cc.Nodes) {
		t.Fatalf("round trip changed topology: %+v, %v", restart, err)
	}
	restored.Nodes = restart
	updated, _, err := configureNodes(restored, &cc)
	if err != nil || !reflect.DeepEqual(updated.Nodes, cc.Nodes) {
		t.Fatalf("restart discarded saved guest metadata: %+v, %v", updated.Nodes, err)
	}
}
