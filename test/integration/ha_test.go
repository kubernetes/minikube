//go:build integration

/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/util/retry"
)

// TestMultiControlPlane tests all ha (multi-control plane) cluster functionality
func TestMultiControlPlane(t *testing.T) {
	if NoneDriver() {
		t.Skip("none driver does not support multinode/ha(multi-control plane) cluster")
	}

	if DockerDriver() {
		rr, err := Run(t, exec.Command("docker", "version", "-f", "{{.Server.Version}}"))
		if err != nil {
			t.Fatalf("docker is broken: %v", err)
		}
		if strings.Contains(rr.Stdout.String(), "azure") {
			t.Skip("kic containers are not supported on docker's azure")
		}
	}

	type validatorFunc func(context.Context, *testing.T, string)
	profile := UniqueProfileName("ha")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validatorFunc
		}{
			{"StartCluster", validateHAStartCluster},
			{"DeployApp", validateHADeployApp},
			{"PingHostFromPods", validateHAPingHostFromPods},
			{"AddWorkerNode", validateHAAddWorkerNode},
			{"NodeLabels", validateHANodeLabels},
			{"HAppyAfterClusterStart", validateHAStatusHAppy},
			{"CopyFile", validateHACopyFile},
			{"StopSecondaryNode", validateHAStopSecondaryNode},
			{"DegradedAfterControlPlaneNodeStop", validateHAStatusDegraded},
			{"RestartSecondaryNode", validateHARestartSecondaryNode},
			{"HAppyAfterSecondaryNodeRestart", validateHAStatusHAppy},
			{"RestartClusterKeepsNodes", validateHARestartClusterKeepsNodes},
			{"DeleteSecondaryNode", validateHADeleteSecondaryNode},
			{"DegradedAfterSecondaryNodeDelete", validateHAStatusDegraded},
			{"StopCluster", validateHAStopCluster},
			{"RestartCluster", validateHARestartCluster},
			{"DegradedAfterClusterRestart", validateHAStatusDegraded},
			{"AddSecondaryNode", validateHAAddSecondaryNode},
			{"HAppyAfterSecondaryNodeAdd", validateHAStatusHAppy},
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

// validateHAStartCluster ensures ha (multi-control plane) cluster can start.
func validateHAStartCluster(ctx context.Context, t *testing.T, profile string) {
	// start ha (multi-control plane) cluster
	startArgs := append([]string{"start", "-p", profile, "--wait=true", "--memory=2200", "--ha", "-v=7", "--alsologtostderr"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to fresh-start ha (multi-control plane) cluster. args %q : %v", rr.Command(), err)
	}

	// ensure minikube status shows 3 operational control-plane nodes
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "type: Control Plane") != 3 {
		t.Errorf("status says not all three control-plane nodes are present: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "host: Running") != 3 {
		t.Errorf("status says not all three hosts are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 3 {
		t.Errorf("status says not all three kubelets are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "apiserver: Running") != 3 {
		t.Errorf("status says not all three apiservers are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
}

// validateHADeployApp deploys an app to ha (multi-control plane) cluster and ensures all nodes can serve traffic.
func validateHADeployApp(ctx context.Context, t *testing.T, profile string) {
	// Create a deployment for app
	_, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "apply", "-f", "./testdata/ha/ha-pod-dns-test.yaml"))
	if err != nil {
		t.Errorf("failed to create busybox deployment to ha (multi-control plane) cluster")
	}

	_, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "rollout", "status", "deployment/busybox"))
	if err != nil {
		t.Errorf("failed to deploy busybox to ha (multi-control plane) cluster")
	}

	// resolve Pod IPs
	resolvePodIPs := func() error {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "get", "pods", "-o", "jsonpath='{.items[*].status.podIP}'"))
		if err != nil {
			err := fmt.Errorf("failed to retrieve Pod IPs (may be temporary): %v", err)
			t.Log(err.Error())
			return err
		}
		podIPs := strings.Split(strings.Trim(rr.Stdout.String(), "'"), " ")
		if len(podIPs) != 3 {
			err := fmt.Errorf("expected 3 Pod IPs but got %d (may be temporary), output: %q", len(podIPs), rr.Output())
			t.Log(err.Error())
			return err
		} else if podIPs[0] == podIPs[1] || podIPs[0] == podIPs[2] || podIPs[1] == podIPs[2] {
			err := fmt.Errorf("expected 3 different pod IPs but got %s and %s (may be temporary), output: %q", podIPs[0], podIPs[1], rr.Output())
			t.Log(err.Error())
			return err
		}
		return nil
	}
	if err := retry.Expo(resolvePodIPs, 1*time.Second, Seconds(120)); err != nil {
		t.Errorf("failed to resolve pod IPs: %v", err)
	}

	// get Pod names
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "get", "pods", "-o", "jsonpath='{.items[*].metadata.name}'"))
	if err != nil {
		t.Errorf("failed get Pod names")
	}
	podNames := strings.Split(strings.Trim(rr.Stdout.String(), "'"), " ")

	// verify all Pods could resolve a public DNS
	for _, name := range podNames {
		_, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "nslookup", "kubernetes.io"))
		if err != nil {
			t.Errorf("Pod %s could not resolve 'kubernetes.io': %v", name, err)
		}
	}

	// verify all Pods could resolve "kubernetes.default"
	// this one is also checked by k8s e2e node conformance tests:
	// https://github.com/kubernetes/kubernetes/blob/f137c4777095b3972e2dd71a01365d47be459389/test/e2e_node/environment/conformance.go#L125-L179
	for _, name := range podNames {
		_, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "nslookup", "kubernetes.default"))
		if err != nil {
			t.Errorf("Pod %s could not resolve 'kubernetes.default': %v", name, err)
		}
	}

	// verify all pods could resolve to a local service.
	for _, name := range podNames {
		_, err = Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "nslookup", "kubernetes.default.svc.cluster.local"))
		if err != nil {
			t.Errorf("Pod %s could not resolve local service (kubernetes.default.svc.cluster.local): %v", name, err)
		}
	}
}

// validateHAPingHostFromPods uses app previously deplyed by validateDeployAppToHACluster to verify its pods, located on different nodes, can resolve "host.minikube.internal".
func validateHAPingHostFromPods(ctx context.Context, t *testing.T, profile string) {
	// get Pod names
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "get", "pods", "-o", "jsonpath='{.items[*].metadata.name}'"))
	if err != nil {
		t.Fatalf("failed to get Pod names: %v", err)
	}
	podNames := strings.Split(strings.Trim(rr.Stdout.String(), "'"), " ")

	for _, name := range podNames {
		// get host.minikube.internal ip as resolved by nslookup
		out, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "sh", "-c", "nslookup host.minikube.internal | awk 'NR==5' | cut -d' ' -f3"))
		if err != nil {
			t.Errorf("Pod %s could not resolve 'host.minikube.internal': %v", name, err)
			continue
		}
		hostIP := net.ParseIP(strings.TrimSpace(out.Stdout.String()))
		if hostIP == nil {
			t.Fatalf("minikube host ip is nil: %s", out.Output())
		}
		// try pinging host from pod
		ping := fmt.Sprintf("ping -c 1 %s", hostIP)
		if _, err := Run(t, exec.CommandContext(ctx, Target(), "kubectl", "-p", profile, "--", "exec", name, "--", "sh", "-c", ping)); err != nil {
			t.Errorf("Failed to ping host (%s) from pod (%s): %v", hostIP, name, err)
		}
	}
}

// validateHAAddWorkerNode uses the minikube node add command to add a worker node to an existing ha (multi-control plane) cluster.
func validateHAAddWorkerNode(ctx context.Context, t *testing.T, profile string) {
	// add a node to the current ha (multi-control plane) cluster
	addArgs := []string{"node", "add", "-p", profile, "-v=7", "--alsologtostderr"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), addArgs...))
	if err != nil {
		t.Fatalf("failed to add worker node to current ha (multi-control plane) cluster. args %q : %v", rr.Command(), err)
	}

	// ensure minikube status shows 3 operational control-plane nodes and 1 worker node
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "type: Control Plane") != 3 {
		t.Errorf("status says not all three control-plane nodes are present: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "host: Running") != 4 {
		t.Errorf("status says not all four hosts are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 4 {
		t.Errorf("status says not all four kubelets are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "apiserver: Running") != 3 {
		t.Errorf("status says not all three apiservers are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
}

// validateHANodeLabels check if all node labels were configured correctly.
func validateHANodeLabels(ctx context.Context, t *testing.T, profile string) {
	// docs: Get the node labels from the cluster with `kubectl get nodes`
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "nodes", "-o", "jsonpath=[{range .items[*]}{.metadata.labels},{end}]"))
	if err != nil {
		t.Errorf("failed to 'kubectl get nodes' with args %q: %v", rr.Command(), err)
	}

	nodeLabelsList := []map[string]string{}
	fixedString := strings.Replace(rr.Stdout.String(), ",]", "]", 1)
	err = json.Unmarshal([]byte(fixedString), &nodeLabelsList)
	if err != nil {
		t.Errorf("failed to decode json from label list: args %q: %v", rr.Command(), err)
	}

	// docs: check if all node labels matches with the expected Minikube labels: `minikube.k8s.io/*`
	expectedLabels := []string{"minikube.k8s.io/commit", "minikube.k8s.io/version", "minikube.k8s.io/updated_at", "minikube.k8s.io/name", "minikube.k8s.io/primary"}

	for _, nodeLabels := range nodeLabelsList {
		for _, el := range expectedLabels {
			if _, ok := nodeLabels[el]; !ok {
				t.Errorf("expected to have label %q in node labels but got : %s", el, rr.Output())
			}
		}
	}
}

// validateHAStatusHAppy ensures minikube profile list outputs correct with ha (multi-control plane) clusters.
func validateHAStatusHAppy(ctx context.Context, t *testing.T, profile string) {
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
	} else {
		if expected, numNodes := 4, len(profileObject.Config.Nodes); numNodes != expected {
			t.Errorf("expected profile %q in json of 'profile list' to include %d nodes but have %d nodes. got *%q*. args: %q", profile, expected, numNodes, rr.Stdout.String(), rr.Command())
		}

		if expected, status := "HAppy", profileObject.Status; status != expected {
			t.Errorf("expected profile %q in json of 'profile list' to have %q status but have %q status. got *%q*. args: %q", profile, expected, status, rr.Stdout.String(), rr.Command())
		}
	}

	if invalidPs, ok := jsonObject["invalid"]; ok {
		for _, ps := range invalidPs {
			if strings.Contains(ps.Name, profile) {
				t.Errorf("expected the json of 'profile list' to not include profile or node in invalid profile but got *%q*. args: %q", rr.Stdout.String(), rr.Command())
			}
		}
	}
}

// validateHACopyFile ensures minikube cp works with ha (multi-control plane) clusters.
func validateHACopyFile(ctx context.Context, t *testing.T, profile string) {
	if NoneDriver() {
		t.Skipf("skipping: cp is unsupported by none driver")
	}

	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--output", "json", "-v=7", "--alsologtostderr"))
	if err != nil && rr.ExitCode != 7 {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	var statuses []cluster.Status
	if err = json.Unmarshal(rr.Stdout.Bytes(), &statuses); err != nil {
		t.Errorf("failed to decode json from status: args %q: %v", rr.Command(), err)
	}

	tmpDir := t.TempDir()

	srcPath := cpTestLocalPath()
	dstPath := cpTestMinikubePath()

	for _, n := range statuses {
		// copy local to node
		testCpCmd(ctx, t, profile, "", srcPath, n.Name, dstPath)

		// copy back from node to local
		tmpPath := filepath.Join(tmpDir, fmt.Sprintf("cp-test_%s.txt", n.Name))
		testCpCmd(ctx, t, profile, n.Name, dstPath, "", tmpPath)

		// copy node to node
		for _, n2 := range statuses {
			if n.Name == n2.Name {
				continue
			}
			fp := path.Join("/home/docker", fmt.Sprintf("cp-test_%s_%s.txt", n.Name, n2.Name))
			testCpCmd(ctx, t, profile, n.Name, dstPath, n2.Name, fp)
		}
	}
}

// validateHAStopSecondaryNode tests ha (multi-control plane) cluster by stopping a secondary control-plane node using minikube node stop command.
func validateHAStopSecondaryNode(ctx context.Context, t *testing.T, profile string) {
	// run minikube node stop on secondary control-plane node
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "stop", SecondNodeName, "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Errorf("secondary control-plane node stop returned an error. args %q: %v", rr.Command(), err)
	}

	// ensure minikube status shows 3 running nodes and 1 stopped node
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-v=7", "--alsologtostderr"))
	// exit code 7 means a host is stopped, which we are expecting
	if err != nil && rr.ExitCode != 7 {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "type: Control Plane") != 3 {
		t.Errorf("status says not all three control-plane nodes are present: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "host: Running") != 3 {
		t.Errorf("status says not three hosts are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 3 {
		t.Errorf("status says not three kubelets are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "apiserver: Running") != 2 {
		t.Errorf("status says not two apiservers are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
}

// validateHAStatusDegraded ensures minikube profile list outputs correct with ha (multi-control plane) clusters.
func validateHAStatusDegraded(ctx context.Context, t *testing.T, profile string) {
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
	} else if expected, status := "Degraded", profileObject.Status; status != expected {
		t.Errorf("expected profile %q in json of 'profile list' to have %q status but have %q status. got *%q*. args: %q", profile, expected, status, rr.Stdout.String(), rr.Command())
	}
}

// validateHARestartSecondaryNode tests the minikube node start command on existing stopped secondary node.
func validateHARestartSecondaryNode(ctx context.Context, t *testing.T, profile string) {
	// start stopped node(s) back up
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "start", SecondNodeName, "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Log(rr.Stderr.String())
		t.Errorf("secondary control-plane node start returned an error. args %q: %v", rr.Command(), err)
	}

	// ensure minikube status shows all 4 nodes running, waiting for ha (multi-control plane) cluster/apiservers to stabilise
	minikubeStatus := func() error {
		rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-v=7", "--alsologtostderr"))
		return err
	}
	if err := retry.Expo(minikubeStatus, 1*time.Second, 60*time.Second); err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "type: Control Plane") != 3 {
		t.Errorf("status says not all three control-plane nodes are present: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "host: Running") != 4 {
		t.Errorf("status says not all four hosts are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 4 {
		t.Errorf("status says not all four kubelets are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "apiserver: Running") != 3 {
		t.Errorf("status says not all three apiservers are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	// ensure kubectl can connect correctly
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes"))
	if err != nil {
		t.Fatalf("failed to kubectl get nodes. args %q : %v", rr.Command(), err)
	}
}

// validateHARestartClusterKeepsNodes restarts minikube cluster and checks if the reported node list is unchanged.
func validateHARestartClusterKeepsNodes(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "node", "list", "-p", profile, "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Errorf("failed to run node list. args %q : %v", rr.Command(), err)
	}
	nodeList := rr.Stdout.String()

	_, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile, "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Errorf("failed to run minikube stop. args %q : %v", rr.Command(), err)
	}

	_, err = Run(t, exec.CommandContext(ctx, Target(), "start", "-p", profile, "--wait=true", "-v=7", "--alsologtostderr"))
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

// validateHADeleteSecondaryNode tests the minikube node delete command on secondary control-plane.
// note: currently, 'minikube status' subcommand relies on primary control-plane node and storage-provisioner only runs on a primary control-plane node.
func validateHADeleteSecondaryNode(ctx context.Context, t *testing.T, profile string) {
	// delete the other secondary control-plane node
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "node", "delete", ThirdNodeName, "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Errorf("node delete returned an error. args %q: %v", rr.Command(), err)
	}

	// ensure status is back down to 3 hosts
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "type: Control Plane") != 2 {
		t.Errorf("status says not two control-plane nodes are present: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "host: Running") != 3 {
		t.Errorf("status says not three hosts are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 3 {
		t.Errorf("status says not three kubelets are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "apiserver: Running") != 2 {
		t.Errorf("status says not two apiservers are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	// ensure kubectl knows the node is gone
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes"))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "NotReady") > 0 {
		t.Errorf("expected 3 nodes to be Ready, got %v", rr.Output())
	}

	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", `go-template='{{range .items}}{{range .status.conditions}}{{if eq .type "Ready"}} {{.status}}{{"\n"}}{{end}}{{end}}{{end}}'`))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "True") != 3 {
		t.Errorf("expected 3 nodes Ready status to be True, got %v", rr.Output())
	}
}

// validateHAStopCluster runs minikube stop on a ha (multi-control plane) cluster.
func validateHAStopCluster(ctx context.Context, t *testing.T, profile string) {
	// Run minikube stop on the cluster
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "stop", "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Errorf("failed to stop cluster. args %q: %v", rr.Command(), err)
	}

	// ensure minikube status shows all 3 nodes stopped
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-v=7", "--alsologtostderr"))
	// exit code 7 means a host is stopped, which we are expecting
	if err != nil && rr.ExitCode != 7 {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "type: Control Plane") != 2 {
		t.Errorf("status says not two control-plane nodes are present: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "host: Running") != 0 {
		t.Errorf("status says there are running hosts: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "kubelet: Stopped") != 3 {
		t.Errorf("status says not three kubelets are stopped: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "apiserver: Stopped") != 2 {
		t.Errorf("status says not two apiservers are stopped: args %q: %v", rr.Command(), rr.Stdout.String())
	}
}

// validateHARestartCluster verifies a soft restart on a ha (multi-control plane) cluster works.
func validateHARestartCluster(ctx context.Context, t *testing.T, profile string) {
	// restart cluster with minikube start
	startArgs := append([]string{"start", "-p", profile, "--wait=true", "-v=7", "--alsologtostderr"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start cluster. args %q : %v", rr.Command(), err)
	}

	// ensure minikube status shows all 3 nodes running
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "type: Control Plane") != 2 {
		t.Errorf("status says not two control-plane nodes are present: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "host: Running") != 3 {
		t.Errorf("status says not three hosts are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 3 {
		t.Errorf("status says not three kubelets are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "apiserver: Running") != 2 {
		t.Errorf("status says not two apiservers are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}

	// ensure kubectl reports that all nodes are ready
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes"))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "NotReady") > 0 {
		t.Errorf("expected 3 nodes to be Ready, got %v", rr.Output())
	}

	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", `go-template='{{range .items}}{{range .status.conditions}}{{if eq .type "Ready"}} {{.status}}{{"\n"}}{{end}}{{end}}{{end}}'`))
	if err != nil {
		t.Fatalf("failed to run kubectl get nodes. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "True") != 3 {
		t.Errorf("expected 3 nodes Ready status to be True, got %v", rr.Output())
	}
}

// validateHAAddSecondaryNode uses the minikube node add command to add a secondary control-plane node to an existing ha (multi-control plane) cluster.
func validateHAAddSecondaryNode(ctx context.Context, t *testing.T, profile string) {
	// add a node to the current ha (multi-control plane) cluster
	addArgs := []string{"node", "add", "-p", profile, "--control-plane", "-v=7", "--alsologtostderr"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), addArgs...))
	if err != nil {
		t.Fatalf("failed to add control-plane node to current ha (multi-control plane) cluster. args %q : %v", rr.Command(), err)
	}

	// ensure minikube status shows 3 operational control-plane nodes and 1 worker node
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-v=7", "--alsologtostderr"))
	if err != nil {
		t.Fatalf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}
	if strings.Count(rr.Stdout.String(), "type: Control Plane") != 3 {
		t.Errorf("status says not all three control-plane nodes are present: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "host: Running") != 4 {
		t.Errorf("status says not all four hosts are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "kubelet: Running") != 4 {
		t.Errorf("status says not all four kubelets are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
	if strings.Count(rr.Stdout.String(), "apiserver: Running") != 3 {
		t.Errorf("status says not all three apiservers are running: args %q: %v", rr.Command(), rr.Stdout.String())
	}
}
