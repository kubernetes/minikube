/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

// newTwoNodeCluster builds a minimal two-node ClusterConfig suitable for
// profile round-tripping. Fields unrelated to node membership are left zero.
func newTwoNodeCluster(name string) *config.ClusterConfig {
	return &config.ClusterConfig{
		Name: name,
		Nodes: []config.Node{
			{Name: "", IP: "192.168.49.2", Port: 8443, ControlPlane: true},
			{Name: "m02", IP: "192.168.49.3", Port: 0, Worker: true},
		},
	}
}

func TestRetrieve_FindsByName(t *testing.T) {
	cc := *newTwoNodeCluster("retrieve-by-name")

	tests := []struct {
		name      string
		query     string
		wantIndex int
		wantName  string
		wantErr   bool
	}{
		{name: "primary by empty name", query: "", wantIndex: 0, wantName: ""},
		{name: "worker by name", query: "m02", wantIndex: 1, wantName: "m02"},
		{name: "worker by machine name", query: "retrieve-by-name-m02", wantIndex: 1, wantName: "m02"},
		{name: "missing", query: "m99", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n, idx, err := Retrieve(cc, tc.query)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Retrieve(%q): expected error, got node=%+v idx=%d", tc.query, n, idx)
				}
				if !strings.Contains(err.Error(), "Could not find node") {
					t.Fatalf("Retrieve(%q): unexpected error %q", tc.query, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Retrieve(%q): unexpected error: %v", tc.query, err)
			}
			if idx != tc.wantIndex {
				t.Fatalf("Retrieve(%q): index = %d, want %d", tc.query, idx, tc.wantIndex)
			}
			if n == nil || n.Name != tc.wantName {
				t.Fatalf("Retrieve(%q): node = %+v, want name %q", tc.query, n, tc.wantName)
			}
		})
	}
}

func TestRemoveFromProfile_DropsNode(t *testing.T) {
	miniHome := t.TempDir()
	cc := newTwoNodeCluster("remove-drops")

	// Seed the profile on disk.
	if err := config.SaveProfile(cc.Name, cc, miniHome); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	removed, err := removeFromProfile(cc, "m02", miniHome)
	if err != nil {
		t.Fatalf("removeFromProfile: %v", err)
	}
	if removed == nil || removed.Name != "m02" {
		t.Fatalf("removeFromProfile returned %+v, want node m02", removed)
	}
	if got := len(cc.Nodes); got != 1 {
		t.Fatalf("in-memory cc.Nodes = %d, want 1", got)
	}
	if cc.Nodes[0].Name != "" {
		t.Fatalf("surviving node = %q, want primary (empty name)", cc.Nodes[0].Name)
	}

	// Reload from disk and confirm the removal was persisted.
	reloaded, err := config.Load(cc.Name, miniHome)
	if err != nil {
		t.Fatalf("config.Load after remove: %v", err)
	}
	if got := len(reloaded.Nodes); got != 1 {
		t.Fatalf("reloaded Nodes = %d, want 1", got)
	}
	if reloaded.Nodes[0].Name != "" {
		t.Fatalf("reloaded surviving node = %q, want primary", reloaded.Nodes[0].Name)
	}
}

func TestRemoveFromProfile_NotFound(t *testing.T) {
	miniHome := t.TempDir()
	cc := newTwoNodeCluster("remove-missing")

	if err := config.SaveProfile(cc.Name, cc, miniHome); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	removed, err := removeFromProfile(cc, "m99", miniHome)
	if err == nil {
		t.Fatalf("removeFromProfile: expected error for missing node, got removed=%+v", removed)
	}
	if !strings.Contains(err.Error(), "retrieve:") {
		t.Fatalf("removeFromProfile: error %q should be wrapped with %q", err, "retrieve:")
	}
	if got := len(cc.Nodes); got != 2 {
		t.Fatalf("cc.Nodes after failed remove = %d, want 2", got)
	}

	// Profile on disk must be untouched.
	reloaded, err := config.Load(cc.Name, miniHome)
	if err != nil {
		t.Fatalf("config.Load after failed remove: %v", err)
	}
	if got := len(reloaded.Nodes); got != 2 {
		t.Fatalf("reloaded Nodes = %d, want 2 (profile should be untouched)", got)
	}
}
