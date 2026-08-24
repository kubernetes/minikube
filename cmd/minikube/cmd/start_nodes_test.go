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
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/run"
)

func nodeFlagCommand(t *testing.T, args ...string) (*cobra.Command, *viper.Viper) {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().StringArray(nodeSpec, nil, "")
	cmd.Flags().IntP(nodes, "n", 1, "")
	cmd.Flags().Bool(ha, false, "")
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
		t.Fatalf("--node must be a visible stringArray flag: %+v", f)
	}
	if startCmd.Flags().Lookup("node-os") != nil {
		t.Fatal("obsolete --node-os flag is still registered")
	}
}

func TestParseNodeSpecs(t *testing.T) {
	cp := config.Node{ControlPlane: true, Worker: true}
	worker := config.Node{Worker: true}
	win := config.Node{Worker: true, Guest: config.Guest{Name: config.GuestOSWindows}}
	for _, tc := range []struct {
		name  string
		specs []string
		want  []config.Node
		err   string
	}{
		{"control plane only", []string{"role=control-plane"}, []config.Node{cp}, ""},
		{"Linux worker", []string{"role=control-plane", "role=worker,os=linux"}, []config.Node{cp, worker}, ""},
		{"mixed schema", []string{"role=control-plane", "role=worker,os=windows"}, []config.Node{cp, win}, ""},
		{"control plane first", []string{"role=worker,os=windows", "role=control-plane", "role=worker"}, []config.Node{cp, win, worker}, ""},
		{"case and whitespace", []string{" ROLE = Control-Plane , OS = Linux "}, []config.Node{cp}, ""},
		{"missing specifications", nil, nil, "at least one node specification"},
		{"empty specification", []string{""}, nil, "nonempty key=value"},
		{"missing role", []string{"os=linux"}, nil, "role is required"},
		{"missing equals", []string{"control-plane"}, nil, "nonempty key=value"},
		{"empty value", []string{"role="}, nil, "nonempty key=value"},
		{"empty key", []string{"=worker"}, nil, "nonempty key=value"},
		{"trailing comma", []string{"role=control-plane,"}, nil, "nonempty key=value"},
		{"duplicate role", []string{"role=control-plane,role=worker"}, nil, `duplicate key "role"`},
		{"duplicate OS", []string{"role=worker,os=linux,OS=windows"}, nil, `duplicate key "os"`},
		{"unknown key", []string{"role=control-plane,name=custom"}, nil, `unknown key "name"`},
		{"unknown role", []string{"role=master"}, nil, "unsupported role"},
		{"unknown OS", []string{"role=worker,os=darwin"}, nil, "unsupported OS"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseNodeSpecs(tc.specs)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v; want %q", err, tc.err)
				}
			} else if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("nodes = %+v, error = %v; want %+v", got, err, tc.want)
			}
		})
	}
}

func TestResolveNodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want int
		err  string
	}{
		{"single Linux node", []string{"--node", "role=control-plane"}, 1, ""},
		{"worker first", []string{"--node=role=worker", "--node=role=control-plane"}, 2, ""},
		{"multiple Linux workers", []string{"--node=role=control-plane", "--node=role=worker", "--node=role=worker"}, 3, ""},
		{"count conflict", []string{"--node=role=control-plane", "-n", "1"}, 0, "cannot be combined"},
		{"HA conflict", []string{"--node=role=control-plane", "--ha"}, 0, "cannot be combined"},
		{"false HA conflict", []string{"--node=role=control-plane", "--ha=false"}, 0, "cannot be combined"},
		{"empty flag", []string{"--node="}, 0, "at least one node specification"},
		{"empty second flag", []string{"--node=role=control-plane", "--node="}, 0, "nonempty key=value"},
		{"worker only", []string{"--node=role=worker"}, 0, "requires a Linux control-plane"},
		{"Windows worker only", []string{"--node=role=worker,os=windows"}, 0, "requires a Linux control-plane"},
		{"Windows control plane", []string{"--node=role=control-plane,os=windows"}, 0, "Windows control-plane nodes are not supported"},
		{"Windows provisioning deferred", []string{"--node=role=control-plane", "--node=role=worker,os=windows"}, 0, "Windows worker provisioning is not available in this build"},
		{"two control planes", []string{"--node=role=control-plane", "--node=role=control-plane"}, 0, "currently supports one Linux control-plane"},
		{"HA through node deferred", []string{"--node=role=control-plane", "--node=role=control-plane", "--node=role=control-plane"}, 0, "currently supports one Linux control-plane"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, v := nodeFlagCommand(t, tc.args...)
			v.SetDefault(nodes, 5)
			v.Set("driver", "docker")
			v.Set(cniFlag, "calico")
			v.Set(containerRuntime, "docker")
			before := v.AllSettings()
			ns, err := resolveNodes(cmd, nil, v)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v; want %q", err, tc.err)
				}
			} else if err != nil || len(ns) != tc.want || !ns[0].ControlPlane {
				t.Fatalf("nodes = %+v, error = %v; want %d nodes, control plane first", ns, err, tc.want)
			}
			if !reflect.DeepEqual(before, v.AllSettings()) {
				t.Fatal("node resolution changed settings")
			}
			applied, applyErr := applyNodeSpecs(cmd, nil, v)
			if (err == nil) != (applyErr == nil) || !reflect.DeepEqual(applied, ns) {
				t.Fatalf("applying node specifications changed the resolution result: %+v, %v", applied, applyErr)
			}
			if err == nil {
				before[nodes] = tc.want
			}
			if !reflect.DeepEqual(before, v.AllSettings()) {
				t.Fatal("applying node specifications must only update the count after successful validation")
			}
		})
	}
}

func TestNodeLegacyInputsUnchanged(t *testing.T) {
	for _, args := range [][]string{nil, {"-n", "2"}, {"--ha"}, {"--ha", "-n", "5"}, {"--ha", "-n", "2"}} {
		cmd, v := nodeFlagCommand(t, args...)
		before := v.AllSettings()
		for _, existing := range []*config.ClusterConfig{nil, {Nodes: []config.Node{{ControlPlane: true}, {Name: "m02"}}}} {
			ns, err := applyNodeSpecs(cmd, existing, v)
			if err != nil || ns != nil || !reflect.DeepEqual(before, v.AllSettings()) {
				t.Fatalf("legacy input %v was intercepted: %+v, %v", args, ns, err)
			}
		}
	}
}

func TestNodeImplicitHAConflict(t *testing.T) {
	cmd, v := nodeFlagCommand(t, "--node=role=control-plane")
	v.SetDefault(ha, true)
	if _, err := resolveNodes(cmd, nil, v); err == nil || !strings.Contains(err.Error(), "HA to be disabled") {
		t.Fatalf("expected implicit HA conflict, got %v", err)
	}
	if !v.GetBool(ha) {
		t.Fatal("HA setting was silently changed")
	}
}

func TestNodeSpecsAreCLIOnly(t *testing.T) {
	cmd, v := nodeFlagCommand(t)
	v.Set(nodeSpec, []string{"role=control-plane", "role=worker,os=windows"})
	ns, err := resolveNodes(cmd, nil, v)
	if err != nil || ns != nil {
		t.Fatalf("configuration without --node was intercepted: %+v, %v", ns, err)
	}
}

func TestNodeSavedTopology(t *testing.T) {
	saved := config.ClusterConfig{
		Name: "saved",
		Nodes: []config.Node{
			{ControlPlane: true, Worker: true, IP: "192.0.2.1", Port: 8443},
			{Name: "m02", Worker: true, IP: "192.0.2.2", Guest: config.Guest{Name: "linux"}},
		},
	}
	for _, tc := range []struct {
		name string
		args []string
		err  string
	}{
		{"matching topology", []string{"--node=role=worker", "--node=role=control-plane"}, ""},
		{"different count", []string{"--node=role=control-plane"}, "cannot change the topology"},
		{"different OS", []string{"--node=role=control-plane", "--node=role=worker,os=windows"}, "Windows worker provisioning"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, v := nodeFlagCommand(t, tc.args...)
			ns, err := resolveNodes(cmd, &saved, v)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v; want %q", err, tc.err)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(ns, saved.Nodes) {
				t.Fatalf("saved topology changed: %+v, %v", ns, err)
			}
			ns[1].IP = "changed"
			if saved.Nodes[1].IP != "192.0.2.2" {
				t.Fatal("resolved nodes alias saved profile")
			}
		})
	}
	for _, savedNodes := range [][]config.Node{
		nil,
		{{ControlPlane: true, Worker: true}, {Name: "m02", ControlPlane: true, Worker: true}},
		{{ControlPlane: true, Worker: true}, {Name: "m02", Worker: true, Guest: config.Guest{Name: config.GuestOSWindows}}},
	} {
		existing := config.ClusterConfig{Name: "saved", Nodes: savedNodes}
		cmd, v := nodeFlagCommand(t, "--node=role=control-plane", "--node=role=worker")
		if _, err := resolveNodes(cmd, &existing, v); err == nil || !strings.Contains(err.Error(), "cannot change the topology") {
			t.Fatalf("expected saved topology conflict for %+v, got %v", savedNodes, err)
		}
	}
}

func TestConfigureNodeSpecs(t *testing.T) {
	for key, value := range map[string]interface{}{
		config.ProfileName:    "node-specs",
		nodes:                 1,
		ha:                    false,
		cpus:                  2,
		memory:                "2048",
		humanReadableDiskSize: defaultDiskSize,
		kvmNUMACount:          1,
		apiServerPort:         8443,
		kubernetesVersion:     constants.DefaultKubernetesVersion,
		containerRuntime:      constants.Containerd,
		cniFlag:               "auto",
		imageRepository:       "",
		imageMirrorCountry:    "",
	} {
		old := viper.Get(key)
		viper.Set(key, value)
		t.Cleanup(func() { viper.Set(key, old) })
	}
	oldCheckRepository := checkRepository
	checkRepository = checkRepoMock
	t.Cleanup(func() { checkRepository = oldCheckRepository })

	for _, specs := range [][]string{{"role=control-plane"}, {"role=worker", "role=control-plane", "role=worker"}} {
		var args []string
		for _, spec := range specs {
			args = append(args, "--node="+spec)
		}
		cmd, v := nodeFlagCommand(t, args...)
		ns, err := resolveNodes(cmd, nil, v)
		if err != nil {
			t.Fatal(err)
		}
		viper.Set(nodes, len(ns))
		cc, primary, err := generateClusterConfig(cmd, nil, constants.DefaultKubernetesVersion, constants.Containerd, driver.Mock, &run.CommandOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if cc.MultiNodeRequested != (len(ns) > 1) || config.IsHA(cc) {
			t.Fatalf("incorrect cluster topology: %+v", cc)
		}
		cc.Nodes = configureNodeSpecs(ns, primary)
		configured := cc.Nodes
		if len(configured) != len(specs) || !configured[0].ControlPlane || !configured[0].Worker {
			t.Fatalf("incorrect configured topology: %+v", configured)
		}
		if !reflect.DeepEqual(configured[0], primary) {
			t.Fatalf("primary node = %+v; want %+v", configured[0], primary)
		}
		for i, n := range configured {
			if n.Port != primary.Port || n.KubernetesVersion != primary.KubernetesVersion || n.ContainerRuntime != primary.ContainerRuntime || n.Guest.IsWindows() {
				t.Fatalf("incorrect node defaults: %+v", n)
			}
			if i > 0 && (n.ControlPlane || !n.Worker || n.Name == "") {
				t.Fatalf("incorrect worker: %+v", n)
			}
		}
		if len(configured) == 3 && (configured[1].Name != "m02" || configured[2].Name != "m03") {
			t.Fatalf("incorrect generated names: %+v", configured)
		}
		if ns[0].Port != 0 {
			t.Fatal("configuration mutated parsed definitions")
		}
		raw, err := json.Marshal(cc)
		if err != nil {
			t.Fatal(err)
		}
		var restored config.ClusterConfig
		if err := json.Unmarshal(raw, &restored); err != nil || !reflect.DeepEqual(restored.Nodes, configured) {
			t.Fatalf("node definitions did not round trip: %+v, %v", restored.Nodes, err)
		}
		if strings.Contains(string(raw), "NodeOS") {
			t.Fatal("obsolete cluster-wide OS list persisted")
		}
	}
}
