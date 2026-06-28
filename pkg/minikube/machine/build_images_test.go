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

package machine

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

func TestShouldSkipNode(t *testing.T) {
	cp := config.Node{Name: "minikube", ControlPlane: true}
	worker1 := config.Node{Name: "minikube-m02", ControlPlane: false}
	worker2 := config.Node{Name: "minikube-m03", ControlPlane: false}

	tests := []struct {
		description string
		node        config.Node
		machineName string
		allNodes    bool
		nodeName    string
		wantSkip    bool
	}{
		// Case 1: allNodes is true - should never skip any node
		{
			description: "allNodes=true on control-plane node",
			node:        cp,
			machineName: "minikube",
			allNodes:    true,
			nodeName:    "",
			wantSkip:    false,
		},
		{
			description: "allNodes=true on worker node",
			node:        worker1,
			machineName: "minikube-m02",
			allNodes:    true,
			nodeName:    "",
			wantSkip:    false,
		},
		{
			description: "allNodes=true with a specific node name requested",
			node:        worker2,
			machineName: "minikube-m03",
			allNodes:    true,
			nodeName:    "minikube-m02",
			wantSkip:    false,
		},

		// Case 2: allNodes is false, nodeName is empty (default behavior)
		{
			description: "default behavior (nodeName=empty) on control-plane node - should NOT skip",
			node:        cp,
			machineName: "minikube",
			allNodes:    false,
			nodeName:    "",
			wantSkip:    false,
		},
		{
			description: "default behavior (nodeName=empty) on worker node - should skip",
			node:        worker1,
			machineName: "minikube-m02",
			allNodes:    false,
			nodeName:    "",
			wantSkip:    true,
		},

		// Case 3: Specific nodeName requested matching user-friendly node name
		{
			description: "nodeName matches worker node name - should NOT skip",
			node:        worker1,
			machineName: "minikube-m02",
			allNodes:    false,
			nodeName:    "minikube-m02",
			wantSkip:    false,
		},
		{
			description: "nodeName matches control-plane name - should NOT skip",
			node:        cp,
			machineName: "minikube",
			allNodes:    false,
			nodeName:    "minikube",
			wantSkip:    false,
		},
		{
			description: "nodeName does not match worker node name - should skip",
			node:        worker1,
			machineName: "minikube-m02",
			allNodes:    false,
			nodeName:    "minikube-m03",
			wantSkip:    true,
		},

		// Case 4: Specific nodeName requested matching VM machine name
		{
			description: "nodeName matches VM machine name - should NOT skip",
			node:        worker1,
			machineName: "minikube-m02-vm",
			allNodes:    false,
			nodeName:    "minikube-m02-vm",
			wantSkip:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			gotSkip := shouldSkipNode(tc.node, cp, tc.machineName, tc.allNodes, tc.nodeName)
			if gotSkip != tc.wantSkip {
				t.Errorf("shouldSkipNode() got %v; want %v", gotSkip, tc.wantSkip)
			}
		})
	}
}
