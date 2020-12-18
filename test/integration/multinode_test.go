/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package integration

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

type clusterStatus struct {
	running          bool
	totalNodes       int
	wantRunningNodes int
	wantStoppedNodes int

	isAzure bool
}

type validatorFunc func(context.Context, *testing.T, string, *clusterStatus)

func TestMultiNode(t *testing.T) {
	if NoneDriver() {
		t.Skip("none driver does not support multinode")
	}

	profile := UniqueProfileName("multinode")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(45))
	defer CleanupWithLogs(t, profile, cancel)

	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validatorFunc
		}{
			{"FreshStart2Nodes", validateMultiNodeStart},
			// Add worker node
			{"AddNode", validateAddNodeToMultiNode},
			{"StopNode", validateStopRunningNode(ThirdNodeName)},
			{"StartAfterStop", validateStartNodeAfterStop(ThirdNodeName)},
			{"DeleteNode", validateDeleteNodeFromMultiNode(ThirdNodeName, true)},
			// Add control plane node
			{"AddControlPlaneNode", validateAddControlPlaneNodeToMultiNode},
			{"StopControlPlaneNode", validateStopRunningNode(ThirdNodeName)},
			{"StartControlPlaneAfterStop", validateStartNodeAfterStop(ThirdNodeName)},
			{"DeleteControlPlaneNode", validateDeleteNodeFromMultiNode(ThirdNodeName, true)},
			// Test cluster stop && start
			{"StopMultiNode", validateStopMultiNodeCluster},
			{"RestartMultiNode", validateRestartMultiNodeCluster},
		}

		s := &clusterStatus{}

		if DockerDriver() {
			rr, err := Run(t, exec.Command("docker", "version", "-f", "{{.Server.Version}}"))
			if err != nil {
				t.Fatalf("docker is broken: %v", err)
			}
			if strings.Contains(rr.Stdout.String(), "azure") {
				s.isAzure = true
			}
		}

		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				defer PostMortemLogs(t, profile)
				tc.validator(ctx, t, profile, s)
			})
		}
	})
}

func validateMultiNodeStart(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
	// Start a 2 node cluster with the --nodes param
	startArgs := append([]string{"start", "-p", profile, "--wait=true", "--memory=2200", "--nodes=2", "-v=8", "--alsologtostderr"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start cluster. args %q : %v", rr.Command(), err)
	}

	s.startCluster()
	s.addNode(2)
	validateClusterStatus(ctx, t, profile, s)
}

func validateAddNodeToMultiNode(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
	// Add a node to the current cluster
	addArgs := []string{"node", "add", "-p", profile, "-v", "3", "--alsologtostderr"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), addArgs...))
	if err != nil {
		t.Fatalf("failed to add node to current cluster. args %q : %v", rr.Command(), err)
	}
	s.addNode(1)
	validateClusterStatus(ctx, t, profile, s)
}

func validateAddControlPlaneNodeToMultiNode(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
	// Add a node to the current cluster
	addArgs := []string{"node", "add", "-p", profile, "-v", "3", "--alsologtostderr", "--control-plane"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), addArgs...))
	if err != nil {
		t.Fatalf("failed to add control plane node to current cluster. args %q : %v", rr.Command(), err)
	}

	s.addNode(1)
	validateClusterStatus(ctx, t, profile, s)
}

func validateStopRunningNode(nodeName string) validatorFunc {
	return func(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
		// Run minikube node stop on that node
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "stop", nodeName))
		if err != nil {
			t.Errorf("node stop returned an error. args %q: %v", rr.Command(), err)
		}

		s.stopNode()
		validateClusterStatus(ctx, t, profile, s)
	}
}

func validateStartNodeAfterStop(nodeName string) validatorFunc {
	return func(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
		if s.isAzure {
			t.Skip("kic containers are not supported on docker's azure")
		}

		// Start the node back up
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "start", nodeName, "--alsologtostderr"))
		if err != nil {
			t.Logf(rr.Stderr.String())
			t.Errorf("node start returned an error. args %q: %v", rr.Command(), err)
		}

		s.startNode()
		validateClusterStatus(ctx, t, profile, s)
	}
}

func validateStopMultiNodeCluster(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
	// Run minikube stop on the cluster
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "stop"))
	if err != nil {
		t.Errorf("node stop returned an error. args %q: %v", rr.Command(), err)
	}

	s.stopCluster()
	validateClusterStatus(ctx, t, profile, s)
}

func validateRestartMultiNodeCluster(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
	if s.isAzure {
		s.startCluster()
		t.Skip("kic containers are not supported on docker's azure")
	}

	// Restart a full cluster with minikube start
	startArgs := append([]string{"start", "-p", profile, "--wait=true", "-v=8", "--alsologtostderr"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start cluster. args %q : %v", rr.Command(), err)
	}

	s.startCluster()
	validateClusterStatus(ctx, t, profile, s)

	// Make sure kubectl reports that all nodes are ready
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes"))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "NotReady") > 0 {
		t.Errorf("expected %v nodes to be Ready, got %v", s.wantRunningNodes, rr.Output())
	}

	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", `go-template='{{range .items}}{{range .status.conditions}}{{if eq .type "Ready"}} {{.status}}{{"\n"}}{{end}}{{end}}{{end}}'`))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "True") != s.wantRunningNodes {
		t.Errorf("expected %v nodes Ready status to be True, got %v", s.wantRunningNodes, rr.Output())
	}
}

func validateDeleteNodeFromMultiNode(nodeName string, running bool) validatorFunc {
	return func(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
		// Start the node back up
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "delete", nodeName))
		if err != nil {
			t.Errorf("node stop returned an error. args %q: %v", rr.Command(), err)
		}

		if running && !s.isAzure {
			s.deleteRunningNode()
		} else {
			s.deleteStoppedNode()
		}
		validateClusterStatus(ctx, t, profile, s)

		if DockerDriver() {
			rr, err := Run(t, exec.Command("docker", "volume", "ls"))
			if err != nil {
				t.Errorf("failed to run %q : %v", rr.Command(), err)
			}
			if strings.Contains(rr.Stdout.String(), fmt.Sprintf("%s-%s", profile, ThirdNodeName)) {
				t.Errorf("docker volume was not properly deleted: %s", rr.Stdout.String())
			}
		}

		// Make sure kubectl knows the node is gone
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes"))
		if err != nil {
			t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
		}
		if strings.Count(rr.Stdout.String(), "NotReady") > 0 {
			t.Errorf("expected %v nodes to be Ready, got %v", s.wantRunningNodes, rr.Output())
		}

		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", `go-template='{{range .items}}{{range .status.conditions}}{{if eq .type "Ready"}} {{.status}}{{"\n"}}{{end}}{{end}}{{end}}'`))
		if err != nil {
			t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
		}
		if strings.Count(rr.Stdout.String(), "True") != s.wantRunningNodes {
			t.Errorf("expected %v nodes Ready status to be True, got %v", s.wantRunningNodes, rr.Output())
		}
	}
}

func (s *clusterStatus) addNode(count int) {
	s.totalNodes += count
	s.wantRunningNodes += count
}

func (s *clusterStatus) stopNode() {
	s.wantRunningNodes--
	s.wantStoppedNodes++
}

func (s *clusterStatus) startNode() {
	s.wantRunningNodes++
	s.wantStoppedNodes--
}

func (s *clusterStatus) deleteRunningNode() {
	s.totalNodes--
	s.wantRunningNodes--
}

func (s *clusterStatus) deleteStoppedNode() {
	s.totalNodes--
	s.wantStoppedNodes--
}

func (s *clusterStatus) stopCluster() {
	s.running = false
	s.wantRunningNodes = 0
	s.wantStoppedNodes = s.totalNodes
}

func (s *clusterStatus) startCluster() {
	s.running = true
	s.wantRunningNodes = s.totalNodes
	s.wantStoppedNodes = 0
}

// validateClusterStatus validates running/stopped kubelet/host count, check kubectl config and api serve connection.
func validateClusterStatus(ctx context.Context, t *testing.T, profile string, s *clusterStatus) {
	// Make sure minikube status shows expected running nodes and stopped nodes
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--alsologtostderr"))
	if err != nil {
		if s.wantStoppedNodes > 0 || s.isAzure { // If isAzure, the start process skipped, so some hosts are stopped
			// Exit code 7 means one host is stopped, which we are expecting
			if rr.ExitCode != 7 {
				t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
			}
		} else {
			t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
		}
	}
	var count int

	count = strings.Count(rr.Stdout.String(), "kubelet: Running")
	if count != s.wantRunningNodes {
		t.Errorf("incorrect number of running kubelets (want: %v, got %v): args %q: %v", s.wantRunningNodes, count, rr.Command(), rr.Stdout.String())
	}

	count = strings.Count(rr.Stdout.String(), "kubelet: Stopped")
	if count != s.wantStoppedNodes {
		t.Errorf("incorrect number of stopped kubelets (want: %v, got %v): args %q: %v", s.wantStoppedNodes, count, rr.Command(), rr.Stdout.String())
	}

	count = strings.Count(rr.Stdout.String(), "host: Running")
	if count != s.wantRunningNodes {
		t.Errorf("incorrect number of running hosts (want: %v, got %v): args %q: %v", s.wantRunningNodes, count, rr.Command(), rr.Stdout.String())
	}

	count = strings.Count(rr.Stdout.String(), "host: Stopped")
	if count != s.wantStoppedNodes {
		t.Errorf("incorrect number of stopped hosts (want: %v, got %v): args %q: %v", s.wantStoppedNodes, count, rr.Command(), rr.Stdout.String())
	}

	if s.running {
		// Make sure kubectl can connect correctly
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes"))
		if err != nil {
			t.Fatalf("failed to kubectl get nodes. args %q : %v", rr.Command(), err)
		}
	}
}
