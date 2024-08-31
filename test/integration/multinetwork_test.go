//go:build integration

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
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/cmd/minikube/cmd"
	"k8s.io/minikube/pkg/drivers/kic/oci"
)

var extnetNetworkName string
var clusterIPv4 string
var clusterIPv6 string
var extnetIPv4 string
var extnetIPv6 string

// TestContainerIPsMultiNetwork tests minikube with docker driver correctly inferring IPs when multiple networks are attached
func TestContainerIPsMultiNetwork(t *testing.T) {
	t.Logf("running with runtime:%s goos:%s goarch:%s", ContainerRuntime(), runtime.GOOS, runtime.GOARCH)
	if !DockerDriver() {
		t.Skip("skipping: only docker driver supported")
	}

	type validateFunc func(context.Context, *testing.T, string)
	profile := UniqueProfileName("extnet")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))

	extnetNetworkName = fmt.Sprintf("%s-%s", "network-extnet", fmt.Sprintf("%06d", time.Now().UnixNano()%1000000))
	defer func() {
		if *cleanup {
			Cleanup(t, profile, cancel)
			CleanupExtnet(t)
		}
	}()

	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"CreateExtnet", createExtnet},
			{"FreshStart", multinetworkValidateFreshStart},
			{"ConnectExtnet", connectExtnet},
			{"Stop", multinetworkValidateStop},
			{"VerifyStatus", multinetworkValidateStatus},
			{"Start", multinetworkValidateStart},
			{"VerifyNetworks", multinetworkValidateNetworks},
			{"Delete", multinetworkValidateDelete},
			{"VerifyDeletedResources", multinetworkValidateVerifyDeleted},
			{"DeleteExtnet", deleteExtnet},
		}
		for _, tc := range tests {
			tc := tc

			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			if t.Failed() {
				t.Fatalf("Previous test failed, not running dependent tests")
			}

			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
			})
		}
	})
}

// createExtnet creates a docker network
func createExtnet(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	cmd := exec.CommandContext(ctx, "docker", "network", "create", extnetNetworkName)

	result, err := Run(t, cmd)
	if err != nil {
		t.Fatalf("failed to execute 'docker network create', error: %v, output: %s", err, result.Output())
	}
	//	extnetNetworkID := result.Output()
	t.Logf("external network %s created", extnetNetworkName)
}

// connectExtnet connects additional network to the minikube cluster
func connectExtnet(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	cmd := exec.CommandContext(ctx, "docker", "network", "connect", extnetNetworkName, profile)

	result, err := Run(t, cmd)
	if err != nil {
		t.Fatalf("failed to execute 'docker network connect', error: %v, output: %s", err, result.Output())
	}
	if KicDriver() {
		bin := "docker"
		if PodmanDriver() {
			bin = "podman"
		}
		extnetIPv4, extnetIPv6, err = oci.ContainerIPs(bin, profile, extnetNetworkName)
		if err != nil {
			t.Fatalf("failed to execute oci.ContainerIPs, error: %v", err)
		}
		t.Logf("cluster %s was attached to network %s with address %s/%s", profile, extnetNetworkName, extnetIPv4, extnetIPv6)
	}
}

// deleteExtnet removes the external network in docker
func deleteExtnet(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	cmd := exec.CommandContext(ctx, "docker", "network", "rm", extnetNetworkName)

	result, err := Run(t, cmd)
	if err != nil {
		t.Fatalf("failed to execute 'docker network delete', error: %v, output: %s", err, result.Output())
	}
	t.Logf("external network %s deleted", extnetNetworkName)
}

// multinetworkValidateFreshStart just starts a new minikube cluster
func multinetworkValidateFreshStart(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := append([]string{"start", "-p", profile, "--memory=2048", "--install-addons=false", "--wait=all"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
	if KicDriver() {
		bin := "docker"
		if PodmanDriver() {
			bin = "podman"
		}
		clusterIPv4, clusterIPv6, err = oci.ContainerIPs(bin, profile, profile)
		if err != nil {
			t.Fatalf("failed to execute oci.ContainerIPs, error: %v", err)
		}
		t.Logf("cluster %s started with address %s/%s", profile, clusterIPv4, clusterIPv6)
	}
}

// multinetworkValidateStop runs minikube Stop
func multinetworkValidateStop(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := []string{"stop", "-p", profile, "--alsologtostderr", "-v=5"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to stop minikube with args: %q : %v", rr.Command(), err)
	}
}

// multinetworkValidateStart runs minikube start
func multinetworkValidateStart(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := []string{"start", "-p", profile, "--alsologtostderr", "-v=5"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	if KicDriver() {
		bin := "docker"
		if PodmanDriver() {
			bin = "podman"
		}
		ipv4, ipv6, err := oci.ContainerIPs(bin, profile, profile)
		if err != nil {
			t.Fatalf("failed to execute oci.ContainerIPs, error: %v", err)
		}
		if ipv4 != clusterIPv4 {
			t.Fatalf("clusterIPv4 mismatch %s != %s", clusterIPv4, ipv4)
		}
		if ipv6 != clusterIPv6 {
			t.Fatalf("clusterIPv6 mismatch %s != %s", clusterIPv6, ipv6)
		}
		ipv4, ipv6, err = oci.ContainerIPs(bin, profile, extnetNetworkName)
		if err != nil {
			t.Fatalf("failed to execute oci.ContainerIPs, error: %v", err)
		}
		if ipv4 != extnetIPv4 {
			t.Fatalf("extnetIPv4 mismatch %s != %s", extnetIPv4, ipv4)
		}
		if ipv6 != extnetIPv6 {
			t.Fatalf("extnetIPv6 mismatch %s != %s", extnetIPv6, ipv6)
		}
	}
}

// Define a struct to match the necessary parts of the JSON structure.
type DockerInspectForNetworks struct {
	NetworkSettings struct {
		Networks map[string]interface{} `json:"Networks"`
	} `json:"NetworkSettings"`
}

// multinetworkValidateNetworks makes sure that the cluster is attached to the correct networks 
func multinetworkValidateNetworks(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.Command("docker", "inspect", profile))
	if err != nil {
		t.Errorf("failed to list profiles with json format after it was deleted. args %q: %v", rr.Command(), err)
	}

	var inspectOutput []DockerInspectForNetworks
	if err := json.Unmarshal(rr.Stdout.Bytes(), &inspectOutput); err != nil {
		t.Errorf("failed to decode json from profile list: args %q: %v", rr.Command(), err)
	}

	networks := inspectOutput[0].NetworkSettings.Networks
	if len(networks) != 2 {
		t.Errorf("expected container to have exactly two networks attached, it has %d networks.\n", len(networks))
	}
	if networks[profile] == nil {
		t.Errorf("expected network %q to be attached to %q, but it is not", profile, profile)
	}
	if networks[extnetNetworkName] == nil {
		t.Errorf("expected network %q to be attached to %q, but it is not", extnetNetworkName, profile)
	}

}

// multinetworkValidateDelete deletes the cluster
func multinetworkValidateDelete(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := []string{"delete", "-p", profile, "--alsologtostderr", "-v=5"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to delete minikube with args: %q : %v", rr.Command(), err)
	}
}

// multinetworkValidateVerifyDeleted makes sure no left over left after deleting a profile such as containers or volumes
func multinetworkValidateVerifyDeleted(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "--output", "json"))
	if err != nil {
		t.Errorf("failed to list profiles with json format after it was deleted. args %q: %v", rr.Command(), err)
	}

	var jsonObject map[string][]map[string]interface{}
	if err := json.Unmarshal(rr.Stdout.Bytes(), &jsonObject); err != nil {
		t.Errorf("failed to decode json from profile list: args %q: %v", rr.Command(), err)
	}
	validProfiles := jsonObject["valid"]
	profileExists := false
	for _, profileObject := range validProfiles {
		if profileObject["Name"] == profile {
			profileExists = true
			break
		}
	}
	if profileExists {
		t.Errorf("expected the deleted profile %q not to show up in profile list but it does! output: %s . args: %q", profile, rr.Stdout.String(), rr.Command())
	}

	if KicDriver() {
		bin := "docker"
		if PodmanDriver() {
			bin = "podman"
		}
		rr, err := Run(t, exec.CommandContext(ctx, bin, "ps", "-a"))
		if err == nil && strings.Contains(rr.Output(), profile) {
			t.Errorf("expected container %q not to exist in output of %s but it does output: %s.", profile, rr.Command(), rr.Output())
		}

		rr, err = Run(t, exec.CommandContext(ctx, bin, "volume", "inspect", profile))
		if err == nil {
			t.Errorf("expected to see error and volume %q to not exist after deletion but got no error and this output: %s", rr.Command(), rr.Output())
		}

		rr, err = Run(t, exec.CommandContext(ctx, bin, "network", "ls"))
		if err != nil {
			t.Errorf("failed to get list of networks: %v", err)
		}
		if strings.Contains(rr.Output(), profile) {
			t.Errorf("expected network %q to not exist after deletion but contained: %s", profile, rr.Output())
		}
	}

}

// multinetworkValidateStatus makes sure stopped clusters show up in minikube status correctly
func multinetworkValidateStatus(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	statusOutput := runStatusCmd(ctx, t, profile, false)
	var cs cmd.ClusterState
	if err := json.Unmarshal(statusOutput, &cs); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	// verify the status looks as we expect
	if cs.StatusCode != cmd.Stopped {
		t.Fatalf("incorrect status code: %v", cs.StatusCode)
	}
	if cs.StatusName != "Stopped" {
		t.Fatalf("incorrect status name: %v", cs.StatusName)
	}
}

// CleanupExtnet removes the external network in docker, no error on failure, used for Cleanup
func CleanupExtnet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), Seconds(10))
	defer cancel()
	t.Logf("Cleaning up docker network %q ...", extnetNetworkName)

	cmd := exec.CommandContext(ctx, "docker", "network", "rm", extnetNetworkName)
	rr := &RunResult{Args: cmd.Args}
	t.Logf("(dbg) Run:  %v", rr.Command())
	err := cmd.Run()
	if err != nil {
		t.Logf("failed to run docker network rm %v", err)
	}
}
