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

package node

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/run"
)

func TestAddingNodePreservesAddons(t *testing.T) {
	for _, controlPlane := range []bool{false, true} {
		cc := &config.ClusterConfig{
			Name: "autoscaler-test",
			Nodes: []config.Node{
				{Name: "autoscaler-test", ControlPlane: true},
				{Name: "autoscaler-test-m02", ControlPlane: controlPlane, Worker: true},
			},
			Addons: map[string]bool{"cluster-autoscaler": true, "metrics-server": true, "dashboard": false},
		}
		// node.Add skips addon installation on the new node (ExistingAddons=nil).
		// That must not clear cluster-wide flags after an autoscaler scale-out.
		updateNodeAddonConfig(Starter{Cfg: cc, Node: &cc.Nodes[1]}, nil, &run.CommandOptions{})
		if !cc.Addons["cluster-autoscaler"] || !cc.Addons["metrics-server"] || cc.Addons["dashboard"] {
			t.Fatalf("adding a node changed addon settings: %v", cc.Addons)
		}
	}
}

// TestManagedWorkerRestartPreservesIdentity keeps VM resume separate from a new join.
func TestManagedWorkerRestartPreservesIdentity(t *testing.T) {
	for _, tt := range []struct {
		name         string
		existing     bool
		controlPlane bool
		managed      bool
		preserve     bool
	}{
		{"managed-existing-worker", true, false, true, true},
		{"managed-new-worker", false, false, true, false},
		{"unmanaged-existing-worker", true, false, false, false},
		{"managed-control-plane", true, true, true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ClusterConfig{VerifyComponents: map[string]bool{kverify.APIServerWaitKey: true}}
			if tt.managed {
				cfg.ClusterAutoscaler = &config.ClusterAutoscalerConfig{}
			}
			starter := Starter{PreExists: tt.existing, Cfg: cfg, Node: &config.Node{ControlPlane: tt.controlPlane}}
			if got := preserveManagedWorker(starter); got != tt.preserve {
				t.Fatalf("preserve managed identity = %v; want %v", got, tt.preserve)
			}
			ready := readinessConfiguration(starter)
			if ready.VerifyComponents[kverify.NodeReadyKey] != tt.preserve || !ready.VerifyComponents[kverify.APIServerWaitKey] {
				t.Fatalf("resume readiness lost required checks: %v", ready.VerifyComponents)
			}
			if cfg.VerifyComponents[kverify.NodeReadyKey] {
				t.Fatal("readiness mutated the caller's component policy")
			}
			ready.VerifyComponents[kverify.APIServerWaitKey] = false
			if !cfg.VerifyComponents[kverify.APIServerWaitKey] {
				t.Fatal("readiness map aliases the caller")
			}
		})
	}
}
