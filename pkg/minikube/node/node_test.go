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
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/run"
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

// mixedCluster builds a 3-node ClusterConfig with primary + orphan + valid
// worker. Used to regression-test the start-path pruning for
// kubernetes/minikube#9834: if only existing.Nodes is pruned, the add-loop
// in startWithDriver re-persists the orphan via SaveProfile(starter.Cfg).
func mixedCluster(name string) *config.ClusterConfig {
	return &config.ClusterConfig{
		Name: name,
		Nodes: []config.Node{
			{Name: "", IP: "192.168.49.2", Port: 8443, ControlPlane: true},
			{Name: "m02", IP: "", Worker: true}, // orphaned: SaveNode ran, Provision did not
			{Name: "m03", IP: "192.168.49.4", Worker: true},
		},
	}
}

func TestPruneOrphanedNodes_NoOpWhenClean(t *testing.T) {
	miniHome := t.TempDir()
	existing := newTwoNodeCluster("prune-clean")
	inflight := newTwoNodeCluster("prune-clean")

	if err := config.SaveProfile(existing.Name, existing, miniHome); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	dropped, err := PruneOrphanedNodes(existing, inflight, existing.Name, miniHome)
	if err != nil {
		t.Fatalf("PruneOrphanedNodes: %v", err)
	}
	if len(dropped) != 0 {
		t.Fatalf("dropped = %v, want empty for clean profile", dropped)
	}
	if len(existing.Nodes) != 2 || len(inflight.Nodes) != 2 {
		t.Fatalf("Nodes len: existing=%d inflight=%d, want 2/2", len(existing.Nodes), len(inflight.Nodes))
	}
}

func TestPruneOrphanedNodes_MixedCluster(t *testing.T) {
	miniHome := t.TempDir()
	// existing is the on-disk view. inflight is the independent copy start
	// hands to the add-loop as starter.Cfg; they share names but are separate
	// slices, exactly how configureNodes builds them in production.
	existing := mixedCluster("prune-mixed")
	inflight := mixedCluster("prune-mixed")

	if err := config.SaveProfile(existing.Name, existing, miniHome); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	dropped, err := PruneOrphanedNodes(existing, inflight, existing.Name, miniHome)
	if err != nil {
		t.Fatalf("PruneOrphanedNodes: %v", err)
	}
	if len(dropped) != 1 || dropped[0] != "m02" {
		t.Fatalf("dropped = %v, want [m02]", dropped)
	}

	// Both slices must be pruned so the add-loop does not re-persist the
	// orphan via node.Add → SaveNode → SaveProfile(starter.Cfg).
	if len(existing.Nodes) != 2 {
		t.Fatalf("existing.Nodes = %d, want 2", len(existing.Nodes))
	}
	if len(inflight.Nodes) != 2 {
		t.Fatalf("inflight.Nodes = %d, want 2 (starter.Cfg must be synced)", len(inflight.Nodes))
	}
	for _, n := range existing.Nodes {
		if n.Name == "m02" {
			t.Fatalf("existing still contains orphan m02: %+v", n)
		}
	}
	for _, n := range inflight.Nodes {
		if n.Name == "m02" {
			t.Fatalf("inflight still contains orphan m02: %+v", n)
		}
	}

	// Disk must match memory: load the profile back and check Nodes.
	reloaded, err := config.Load(existing.Name, miniHome)
	if err != nil {
		t.Fatalf("config.Load after prune: %v", err)
	}
	if len(reloaded.Nodes) != 2 {
		t.Fatalf("reloaded Nodes = %d, want 2", len(reloaded.Nodes))
	}
	for _, n := range reloaded.Nodes {
		if n.Name == "m02" {
			t.Fatalf("on-disk profile still contains orphan m02: %+v", n)
		}
	}
}

func TestPruneOrphanedNodes_PreservesPrimaryWithEmptyIP(t *testing.T) {
	// Defensive: if the primary somehow ended up with an empty IP (shouldn't
	// happen, Provision sets it before worker-add runs), the primary is still
	// preserved. IsPrimaryControlPlane is matched by the name of Nodes[0], so
	// the predicate relies on ordering rather than IP presence.
	miniHome := t.TempDir()
	cc := &config.ClusterConfig{
		Name: "prune-primary",
		Nodes: []config.Node{
			{Name: "", IP: "", Port: 8443, ControlPlane: true},
			{Name: "m02", IP: "", Worker: true},
		},
	}
	if err := config.SaveProfile(cc.Name, cc, miniHome); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	dropped, err := PruneOrphanedNodes(cc, nil, cc.Name, miniHome)
	if err != nil {
		t.Fatalf("PruneOrphanedNodes: %v", err)
	}
	if len(dropped) != 1 || dropped[0] != "m02" {
		t.Fatalf("dropped = %v, want [m02]", dropped)
	}
	if len(cc.Nodes) != 1 || cc.Nodes[0].Name != "" {
		t.Fatalf("cc.Nodes = %+v, want only primary", cc.Nodes)
	}
}

func TestPruneOrphanedNodes_NilExisting(t *testing.T) {
	dropped, err := PruneOrphanedNodes(nil, nil, "no-profile")
	if err != nil {
		t.Fatalf("PruneOrphanedNodes(nil, nil): %v", err)
	}
	if len(dropped) != 0 {
		t.Fatalf("dropped = %v, want empty for nil existing", dropped)
	}
}

func TestDelete_OrphanShortCircuit(t *testing.T) {
	// Redirect config.SaveProfile / config.Load to a temp dir via MINIKUBE_HOME.
	// Delete → removeFromProfile does not take a miniHome, so it falls through
	// to localpath.MiniPath(), which honors the env var set below.
	miniHome := t.TempDir()
	t.Setenv(localpath.MinikubeHome, miniHome)

	cc := &config.ClusterConfig{
		Name: "orphan-delete",
		Driver: "docker",
		Nodes: []config.Node{
			{Name: "", IP: "192.168.49.2", Port: 8443, ControlPlane: true},
			{Name: "m02", IP: "", Worker: true}, // orphan: IP never set
		},
	}
	if err := config.SaveProfile(cc.Name, cc); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	// &run.CommandOptions{} is enough: machine.NewAPIClient never errors, and
	// machine.DeleteHost reports host-not-exists for a store that has no
	// backing files; Delete swallows that into klog and proceeds to the
	// profile update, which is what we assert on.
	n, err := Delete(*cc, "m02", &run.CommandOptions{})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if n == nil || n.Name != "m02" {
		t.Fatalf("Delete returned %+v, want node m02", n)
	}

	reloaded, err := config.Load(cc.Name)
	if err != nil {
		t.Fatalf("config.Load after Delete: %v", err)
	}
	if got := len(reloaded.Nodes); got != 1 {
		t.Fatalf("reloaded Nodes = %d, want 1", got)
	}
	if reloaded.Nodes[0].Name != "" {
		t.Fatalf("surviving node = %q, want primary (empty name)", reloaded.Nodes[0].Name)
	}
}
