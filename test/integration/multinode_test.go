// +build integration

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
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"

	"k8s.io/minikube/cmd/minikube/cmd"
	"k8s.io/minikube/pkg/minikube/config"
)

// TestMultiNode tests all multi node cluster functionality
func TestMultiNode(t *testing.T) {
	if NoneDriver() {
		t.Skip("none driver does not support multinode")
	}

	type validatorFunc func(context.Context, *testing.T, string)
	profile := UniqueProfileName("multinode")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validatorFunc
		}{
			{"FreshStart2Nodes", validateMultiNodeStart},
			{"DeployApp2Nodes", validateDeployAppToMultiNode},
			{"PingHostFrom2Pods", validatePodsPingHost},
			{"AddNode", validateAddNodeToMultiNode},
			{"ProfileList", validateProfileListWithMultiNode},
			{"CopyFile", validateCopyFileWithMultiNode},
			{"StopNode", validateStopRunningNode},
			{"StartAfterStop", validateStartNodeAfterStop},
			{"RestartKeepsNodes", validateRestartKeepsNodes},
			{"DeleteNode", validateDeleteNodeFromMultiNode},
			{"StopMultiNode", validateStopMultiNodeCluster},
			{"RestartMultiNode", validateRestartMultiNodeCluster},
			{"ValidateNameConflict", validateNameConflict},
		}
		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				defer PostMortemLogs(t, profile)
				tc.validator(ctx, t, profile)
			})
		}
	})
}

// validateMultiNodeStart makes sure a 2 node cluster can start
func validateMultiNodeStart(ctx context.Context, t *testing.T, profile string) {
	// Start a 2 node cluster with the --nodes param
	startArgs := append([]string{"start", "-p", profile, "--wait=true", "--memory=2200", "--nodes=2", "-v=8", "--alsologtostderr"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start cluster. args %q : %v", rr.Command(), err)
	}

	// Make sure minikube status shows 2 nodes
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	if strings.Count(rr.Stdout.String(), "host: Running") != 2 {
		t.Errorf("status says both hosts are not running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 2 {
		t.Errorf("status says both kubelets are not running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

}

// validateAddNodeToMultiNode uses the minikube node add command to add a node to an existing cluster
func validateAddNodeToMultiNode(ctx context.Context, t *testing.T, profile string) {
	// Add a node to the current cluster
	addArgs := []string{"node", "add", "-p", profile, "-v", "3", "--alsologtostderr"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), addArgs...))
	if err != nil {
		t.Fatalf("failed to add node to current cluster. args %q : %v", rr.Command(), err)
	}

	// Make sure minikube status shows 3 nodes
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	if strings.Count(rr.Stdout.String(), "host: Running") != 3 {
		t.Errorf("status says all hosts are not running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 3 {
		t.Errorf("status says all kubelets are not running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
}

// validateProfileListWithMultiNode make sure minikube profile list outputs correct with multinode clusters
func validateProfileListWithMultiNode(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "--output", "json"))
	if err != nil {
		t.Errorf("failed to list profiles with json format. args %q: %v", rr.Command(), err)
	}

	var jsonObject map[string][]config.Profile
	err = json.Unmarshal(rr.Stdout.Bytes(), &jsonObject)
	if err != nil {
		t.Errorf("failed to decode json from profile list: args %q: %v", rr.Command(), err)
	}

	validProfiles := jsonObject["valid"]
	var profileObject *config.Profile
	for _, obj := range validProfiles {
		if obj.Name == profile {
			profileObject = &obj
			break
		}
	}

	if profileObject == nil {
		t.Errorf("expected the json of 'profile list' to include %q but got *%q*. args: %q", profile, rr.Stdout.String(), rr.Command())
	} else if expected, numNodes := 3, len(profileObject.Config.Nodes); expected != numNodes {
		t.Errorf("expected profile %q in json of 'profile list' include %d nodes but have %d nodes. got *%q*. args: %q", profile, expected, numNodes, rr.Stdout.String(), rr.Command())
	}

	if invalidPs, ok := jsonObject["invalid"]; ok {
		for _, ps := range invalidPs {
			if strings.Contains(ps.Name, profile) {
				t.Errorf("expected the json of 'profile list' to not include profile or node in invalid profile but got *%q*. args: %q", rr.Stdout.String(), rr.Command())
			}
		}
	}
}

// validateProfileListWithMultiNode make sure minikube profile list outputs correct with multinode clusters
func validateCopyFileWithMultiNode(ctx context.Context, t *testing.T, profile string) {
	if NoneDriver() {
		t.Skipf("skipping: cp is unsupported by none driver")
	}

	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--output", "json", "--alsologtostderr"))
	if err != nil && rr.ExitCode != 7 {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	var statuses []cmd.Status
	if err = json.Unmarshal(rr.Stdout.Bytes(), &statuses); err != nil {
		t.Errorf("failed to decode json from status: args %q: %v", rr.Command(), err)
	}

	for _, s := range statuses {
		if s.Worker {
			testCpCmd(ctx, t, profile, s.Name)
		} else {
			testCpCmd(ctx, t, profile, "")
		}
	}
}

// validateStopRunningNode tests the minikube node stop command
func validateStopRunningNode(ctx context.Context, t *testing.T, profile string) {
	// Run minikube node stop on that node
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "stop", ThirdNodeName))
	if err != nil {
		t.Errorf("node stop returned an error. args %q: %v", rr.Command(), err)
	}

	// Run status again to see the stopped host
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status"))
	// Exit code 7 means one host is stopped, which we are expecting
	if err != nil && rr.ExitCode != 7 {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	// Make sure minikube status shows 2 running nodes and 1 stopped one
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--alsologtostderr"))
	if err != nil && rr.ExitCode != 7 {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 2 {
		t.Errorf("incorrect number of running kubelets: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	if strings.Count(rr.Stdout.String(), "host: Stopped") != 1 {
		t.Errorf("incorrect number of stopped hosts: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	if strings.Count(rr.Stdout.String(), "kubelet: Stopped") != 1 {
		t.Errorf("incorrect number of stopped kubelets: args %q: %v", rr.Command(), rr.Stdout.String())
	}
}

// validateStartNodeAfterStop tests the minikube node start command on an existing stopped node
func validateStartNodeAfterStop(ctx context.Context, t *testing.T, profile string) {
	if DockerDriver() {
		rr, err := Run(t, exec.Command("docker", "version", "-f", "{{.Server.Version}}"))
		if err != nil {
			t.Fatalf("docker is broken: %v", err)
		}
		if strings.Contains(rr.Stdout.String(), "azure") {
			t.Skip("kic containers are not supported on docker's azure")
		}
	}

	// Start the node back up
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "start", ThirdNodeName, "--alsologtostderr"))
	if err != nil {
		t.Logf(rr.Stderr.String())
		t.Errorf("node start returned an error. args %q: %v", rr.Command(), err)
	}

	// Make sure minikube status shows 3 running hosts
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	if strings.Count(rr.Stdout.String(), "host: Running") != 3 {
		t.Errorf("status says both hosts are not running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 3 {
		t.Errorf("status says both kubelets are not running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	// Make sure kubectl can connect correctly
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes"))
	if err != nil {
		t.Fatalf("failed to kubectl get nodes. args %q : %v", rr.Command(), err)
	}
}

// validateRestartKeepsNodes restarts minikube cluster and checks if the reported node list is unchanged
func validateRestartKeepsNodes(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "node", "list", "-p", profile))
	if err != nil {
		t.Errorf("failed to run node list. args %q : %v", rr.Command(), err)
	}

	nodeList := rr.Stdout.String()

	_, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile))
	if err != nil {
		t.Errorf("failed to run minikube stop. args %q : %v", rr.Command(), err)
	}

	_, err = Run(t, exec.CommandContext(ctx, Target(), "start", "-p", profile, "--wait=true", "-v=8", "--alsologtostderr"))
	if err != nil {
		t.Errorf("failed to run minikube start. args %q : %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "node", "list", "-p", profile))
	if err != nil {
		t.Errorf("failed to run node list. args %q : %v", rr.Command(), err)
	}

	restartedNodeList := rr.Stdout.String()
	if nodeList != restartedNodeList {
		t.Fatalf("reported node list is not the same after restart. Before restart: %s\nAfter restart: %s", nodeList, restartedNodeList)
	}
}

// validateStopMultiNodeCluster runs minikube stop on a multinode cluster
func validateStopMultiNodeCluster(ctx context.Context, t *testing.T, profile string) {
	// Run minikube stop on the cluster
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "stop"))
	if err != nil {
		t.Errorf("node stop returned an error. args %q: %v", rr.Command(), err)
	}

	// Run status to see the stopped hosts
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status"))
	// Exit code 7 means one host is stopped, which we are expecting
	if err != nil && rr.ExitCode != 7 {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	// Make sure minikube status shows 2 stopped nodes
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--alsologtostderr"))
	if err != nil && rr.ExitCode != 7 {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	if strings.Count(rr.Stdout.String(), "host: Stopped") != 2 {
		t.Errorf("incorrect number of stopped hosts: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	if strings.Count(rr.Stdout.String(), "kubelet: Stopped") != 2 {
		t.Errorf("incorrect number of stopped kubelets: args %q: %v", rr.Command(), rr.Stdout.String())
	}
}

// validateRestartMultiNodeCluster verifies a soft restart on a multinode cluster works
func validateRestartMultiNodeCluster(ctx context.Context, t *testing.T, profile string) {
	if DockerDriver() {
		rr, err := Run(t, exec.Command("docker", "version", "-f", "{{.Server.Version}}"))
		if err != nil {
			t.Fatalf("docker is broken: %v", err)
		}
		if strings.Contains(rr.Stdout.String(), "azure") {
			t.Skip("kic containers are not supported on docker's azure")
		}
	}
	// Restart a full cluster with minikube start
	startArgs := append([]string{"start", "-p", profile, "--wait=true", "-v=8", "--alsologtostderr"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start cluster. args %q : %v", rr.Command(), err)
	}

	// Make sure minikube status shows 2 running nodes
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	if strings.Count(rr.Stdout.String(), "host: Running") != 2 {
		t.Errorf("status says both hosts are not running: args %q: %v", rr.Command(), rr.Output())
	}

	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 2 {
		t.Errorf("status says both kubelets are not running: args %q: %v", rr.Command(), rr.Output())
	}

	// Make sure kubectl reports that all nodes are ready
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes"))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "NotReady") > 0 {
		t.Errorf("expected 2 nodes to be Ready, got %v", rr.Output())
	}

	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", `go-template='{{range .items}}{{range .status.conditions}}{{if eq .type "Ready"}} {{.status}}{{"\n"}}{{end}}{{end}}{{end}}'`))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "True") != 2 {
		t.Errorf("expected 2 nodes Ready status to be True, got %v", rr.Output())
	}
}

// validateDeleteNodeFromMultiNode tests the minikube node delete command
func validateDeleteNodeFromMultiNode(ctx context.Context, t *testing.T, profile string) {
	// Start the node back up
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "delete", ThirdNodeName))
	if err != nil {
		t.Errorf("node stop returned an error. args %q: %v", rr.Command(), err)
	}

	// Make sure status is back down to 2 hosts
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	if strings.Count(rr.Stdout.String(), "host: Running") != 2 {
		t.Errorf("status says both hosts are not running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 2 {
		t.Errorf("status says both kubelets are not running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

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
		t.Errorf("expected 2 nodes to be Ready, got %v", rr.Output())
	}

	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", `go-template='{{range .items}}{{range .status.conditions}}{{if eq .type "Ready"}} {{.status}}{{"\n"}}{{end}}{{end}}{{end}}'`))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "True") != 2 {
		t.Errorf("expected 2 nodes Ready status to be True, got %v", rr.Output())
	}
}

// validateNameConflict tests that the node name verification works as expected
func validateNameConflict(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "node", "list", "-p", profile))
	if err != nil {
		t.Errorf("failed to run node list. args %q : %v", rr.Command(), err)
	}
	curNodeNum := strings.Count(rr.Stdout.String(), profile)

	// Start new profile. It's expected failture
	profileName := fmt.Sprintf("%s-m0%d", profile, curNodeNum)
	startArgs := append([]string{"start", "-p", profileName}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err == nil {
		t.Errorf("expected start profile command to fail. args %q", rr.Command())
	}

	// Start new profile temporary profile to conflict node name.
	profileName = fmt.Sprintf("%s-m0%d", profile, curNodeNum+1)
	startArgs = append([]string{"start", "-p", profileName}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Errorf("failed to start profile. args %q : %v", rr.Command(), err)
	}

	// Add a node to the current cluster. It's expected failture
	addArgs := []string{"node", "add", "-p", profile}
	rr, err = Run(t, exec.CommandContext(ctx, Target(), addArgs...))
	if err == nil {
		t.Errorf("expected add node command to fail. args %q : %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "delete", "-p", profileName))
	if err != nil {
		t.Logf("failed to clean temporary profile. args %q : %v", rr.Command(), err)
	}
}

// validateDeployAppToMultiNode deploys an app to a multinode cluster and makes sure all nodes can serve traffic
func validateDeployAppToMultiNode(ctx context.Context, t *testing.T, profile string) {
	// Create a deployment for app
	_, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "apply", "-f", "./testdata/multinodes/multinode-pod-dns-test.yaml"))
	if err != nil {
		t.Errorf("failed to create busybox deployment to multinode cluster")
	}

	_, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "rollout", "status", "deployment/busybox"))
	if err != nil {
		t.Errorf("failed to deploy busybox to multinode cluster")
	}

	// resolve Pod IPs
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "get", "pods", "-o", "jsonpath='{.items[*].status.podIP}'"))
	if err != nil {
		t.Errorf("failed to retrieve Pod IPs")
	}
	podIPs := strings.Split(strings.Trim(rr.Stdout.String(), "'"), " ")
	if len(podIPs) != 2 {
		t.Errorf("expected 2 Pod IPs but got %d", len(podIPs))
	} else if podIPs[0] == podIPs[1] {
		t.Errorf("expected 2 different pod IPs but got %s and %s", podIPs[0], podIPs[0])
	}

	// get Pod names
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "get", "pods", "-o", "jsonpath='{.items[*].metadata.name}'"))
	if err != nil {
		t.Errorf("failed get Pod names")
	}
	podNames := strings.Split(strings.Trim(rr.Stdout.String(), "'"), " ")

	// verify both Pods could resolve a public DNS
	for _, name := range podNames {
		_, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "nslookup", "kubernetes.io"))
		if err != nil {
			t.Errorf("Pod %s could not resolve 'kubernetes.io': %v", name, err)
		}
	}

	// verify both Pods could resolve "kubernetes.default"
	// this one is also checked by k8s e2e node conformance tests:
	// https://github.com/kubernetes/kubernetes/blob/f137c4777095b3972e2dd71a01365d47be459389/test/e2e_node/environment/conformance.go#L125-L179
	for _, name := range podNames {
		_, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "nslookup", "kubernetes.default"))
		if err != nil {
			t.Errorf("Pod %s could not resolve 'kubernetes.default': %v", name, err)
		}
	}

	// verify both pods could resolve to a local service.
	for _, name := range podNames {
		_, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "nslookup", "kubernetes.default.svc.cluster.local"))
		if err != nil {
			t.Errorf("Pod %s could not resolve local service (kubernetes.default.svc.cluster.local): %v", name, err)
		}
	}
}

// validatePodsPingHost uses app previously deplyed by validateDeployAppToMultiNode to verify its pods, located on different nodes, can resolve "host.minikube.internal".
func validatePodsPingHost(ctx context.Context, t *testing.T, profile string) {
	// get Pod names
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "get", "pods", "-o", "jsonpath='{.items[*].metadata.name}'"))
	if err != nil {
		t.Errorf("failed get Pod names")
	}
	podNames := strings.Split(strings.Trim(rr.Stdout.String(), "'"), " ")

	for _, name := range podNames {
		// get host.minikube.internal ip as resolved by nslookup
		if out, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "sh", "-c", "nslookup host.minikube.internal | awk 'NR==5' | cut -d' ' -f3")); err != nil {
			t.Errorf("Pod %s could not resolve 'host.minikube.internal': %v", name, err)
		} else {
			hostIP := net.ParseIP(strings.TrimSpace(out.Stdout.String()))
			// get node's eth0 network
			if out, err := Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "ip -4 -br -o a s eth0 | tr -s ' ' | cut -d' ' -f3")); err != nil {
				t.Errorf("Error getting eth0 IP of node %s: %v", profile, err)
			} else {
				if _, nodeNet, err := net.ParseCIDR(strings.TrimSpace(out.Stdout.String())); err != nil {
					t.Errorf("Error parsing eth0 IP of node %s: %v", profile, err)
					// check if host ip belongs to node's eth0 network
				} else if !nodeNet.Contains(hostIP) {
					t.Errorf("Pod %s resolved 'host.minikube.internal' to %s while node's eth0 network is %s", name, hostIP, nodeNet)
				}
			}
		}
	}
}
