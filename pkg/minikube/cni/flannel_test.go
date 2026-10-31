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

package cni

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
)

func TestFlannelManifest(t *testing.T) {
	linux := strings.ReplaceAll(flannelYaml, "{{ .PodCIDR }}", DefaultPodCIDR)
	for _, tc := range []struct {
		name    string
		nodes   []config.Node
		windows bool
	}{
		{name: "legacy Linux"},
		{name: "explicit Linux", nodes: []config.Node{{ControlPlane: true}, {Guest: config.Guest{Name: "linux"}}}},
		{name: "mixed OS", nodes: []config.Node{{ControlPlane: true}, {Guest: config.Guest{Name: config.GuestOSWindows}}}, windows: true},
		{name: "multiple Windows workers", nodes: []config.Node{{ControlPlane: true}, {Guest: config.Guest{Name: config.GuestOSWindows}}, {Guest: config.Guest{Name: config.GuestOSWindows}}}, windows: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := Flannel{cc: config.ClusterConfig{Nodes: tc.nodes}}
			manifest, err := c.manifest()
			if err != nil {
				t.Fatal(err)
			}
			defer manifest.Close()
			data, err := io.ReadAll(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if !tc.windows && string(data) != linux {
				t.Fatal("Linux-only manifest changed")
			}
			if tc.windows && !strings.HasPrefix(string(data), linux) {
				t.Fatal("mixed manifest must preserve Linux resources")
			}
			want := []string{
				"Namespace//kube-flannel",
				"ClusterRole//flannel",
				"ClusterRoleBinding//flannel",
				"ServiceAccount/kube-flannel/flannel",
				"ConfigMap/kube-flannel/kube-flannel-cfg",
				"DaemonSet/kube-flannel/kube-flannel-ds",
			}
			if tc.windows {
				want = append(want,
					"ConfigMap/kube-flannel/kube-flannel-windows-cfg",
					"DaemonSet/kube-flannel/kube-flannel-ds-windows-amd64",
					"DaemonSet/kube-system/kube-proxy-windows")
			}
			var got []string
			decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), 4096)
			for {
				var object metav1.PartialObjectMetadata
				err := decoder.Decode(&object)
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatalf("invalid manifest: %v", err)
				}
				if object.Kind != "" {
					got = append(got, fmt.Sprintf("%s/%s/%s", object.Kind, object.Namespace, object.Name))
				}
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("resources = %v; want %v", got, want)
			}
		})
	}
}

func TestFlannelApply(t *testing.T) {
	for _, windows := range []bool{false, true} {
		t.Run(fmt.Sprintf("windows=%t", windows), func(t *testing.T) {
			cfg := config.ClusterConfig{
				Name: "target-profile",
				KubernetesConfig: config.KubernetesConfig{
					CNI:               "flannel",
					KubernetesVersion: "v1.37.0",
				},
			}
			if windows {
				cfg.Nodes = []config.Node{{ControlPlane: true}, {Guest: config.Guest{Name: config.GuestOSWindows}}}
			}
			runner := &flannelRunner{}
			for range 2 {
				manager, err := New(&cfg)
				if err != nil {
					t.Fatal(err)
				}
				if err := manager.Apply(runner); err != nil {
					t.Fatal(err)
				}
			}
			wantCalls := []string{"stat", "copy", "sudo", "stat", "copy", "sudo"}
			if !reflect.DeepEqual(runner.calls, wantCalls) {
				t.Fatalf("calls = %v; want %v", runner.calls, wantCalls)
			}
			wantApply := []string{"sudo", kapi.KubectlBinaryPath(cfg.KubernetesConfig.KubernetesVersion), "apply", "--kubeconfig=/var/lib/minikube/kubeconfig", "-f", manifestPath()}
			for _, cmd := range runner.commands {
				if cmd[0] == "sudo" && !reflect.DeepEqual(cmd, wantApply) {
					t.Errorf("apply command = %v; want %v", cmd, wantApply)
				}
			}
			if len(runner.manifests) != 2 || runner.manifests[0] != runner.manifests[1] {
				t.Fatal("reapplying CNI must use the same complete manifest")
			}
			for _, name := range []string{"kube-flannel-ds-windows-amd64", "kube-proxy-windows"} {
				if strings.Contains(runner.manifests[0], name) != windows {
					t.Errorf("Windows resource %q present = %t; want %t", name, !windows, windows)
				}
			}
		})
	}
}

func TestFlannelApplyErrors(t *testing.T) {
	failure := errors.New("runner failure")
	for _, tc := range []struct {
		stage string
		calls []string
	}{
		{stage: "stat", calls: []string{"stat"}},
		{stage: "copy", calls: []string{"stat", "copy"}},
		{stage: "sudo", calls: []string{"stat", "copy", "sudo"}},
	} {
		t.Run(tc.stage, func(t *testing.T) {
			runner := &flannelRunner{failStage: tc.stage, failure: failure}
			c := Flannel{cc: config.ClusterConfig{Nodes: []config.Node{{Guest: config.Guest{Name: config.GuestOSWindows}}}}}
			if err := c.Apply(runner); !errors.Is(err, failure) {
				t.Fatalf("error = %v; want runner failure", err)
			}
			if !reflect.DeepEqual(runner.calls, tc.calls) {
				t.Fatalf("calls = %v; want %v", runner.calls, tc.calls)
			}
		})
	}
}

type flannelRunner struct {
	calls     []string
	commands  [][]string
	manifests []string
	failStage string
	failure   error
}

func (r *flannelRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	r.calls = append(r.calls, cmd.Args[0])
	r.commands = append(r.commands, cmd.Args)
	result := &command.RunResult{Args: cmd.Args}
	if cmd.Args[0] == r.failStage {
		return result, r.failure
	}
	return result, nil
}

func (r *flannelRunner) Copy(file assets.CopyableFile) error {
	r.calls = append(r.calls, "copy")
	if r.failStage == "copy" {
		return r.failure
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	r.manifests = append(r.manifests, string(data))
	return nil
}
