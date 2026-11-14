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
