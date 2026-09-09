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

package cluster

import "testing"

func TestApplyTransientCodeToNodes(t *testing.T) {
	nodes := []NodeState{
		{BaseState: BaseState{Name: "minikube", StatusCode: OK, StatusName: "OK"}},
		{BaseState: BaseState{Name: "minikube-m02", StatusCode: OK, StatusName: "OK"}},
	}

	applyTransientCodeToNodes(nodes, InsufficientStorage)

	for i, n := range nodes {
		if n.StatusCode != InsufficientStorage {
			t.Errorf("nodes[%d].StatusCode = %d, want %d", i, n.StatusCode, InsufficientStorage)
		}
		if n.StatusName != "InsufficientStorage" {
			t.Errorf("nodes[%d].StatusName = %q, want InsufficientStorage", i, n.StatusName)
		}
	}
}
