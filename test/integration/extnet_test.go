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
	"fmt"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/cmd/minikube/cmd"
)

// TestExtNet tests minikube external network functionality
func TestExtNet(t *testing.T) {
	driverArg, driverArgPresent := DriverArg()
	t.Logf("running with runtime:%s DriverArg:%s goos:%s goarch:%s", ContainerRuntime(), driverArg, runtime.GOOS, runtime.GOARCH)
	if driverArgPresent && driverArg != "docker" {
		t.Skip("skipping: only docker driver supported")
	}

	type validateFunc func(context.Context, *testing.T, string, string)
	profile := UniqueProfileName("extnet")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	defer Cleanup(t, profile, cancel)

	containerRuntime := ContainerRuntime()
	switch containerRuntime {
	case "containerd":
	    t.Skip("skipping: access direct test is broken on windows: https://github.com/kubernetes/minikube/issues/8304")
	case "crio":
	    t.Skip("skipping: access direct test is broken on windows: https://github.com/kubernetes/minikube/issues/8304")
	case "docker":
	}

	extnetNetworkName := fmt.Sprintf("%s-%s", "network-extnet", fmt.Sprintf("%06d", time.Now().UnixNano()%1000000))

	// Serial tests
	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"CreateExtnet", createExtnet},
			{"FreshStart", extnetValidateFreshStart},
			{"ConnectExtnet", connectExtnet},
			{"Stop", extnetValidateStop},
			{"VerifyStatus", extnetValidateStatus},
			{"Start", extnetValidateStart},
			{"Delete", extnetValidateDelete},
		//	{"Fail", fail},
			{"VerifyDeletedResources", extnetValidateVerifyDeleted},
		}
		for _, tc := range tests {
			tc := tc

			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			if t.Failed() {
				// t.Fatalf("Previous test failed, not running dependent tests")
				break
			}

			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile, extnetNetworkName)
				if t.Failed() && *postMortemLogs {
					PostMortemLogs(t, profile)
				}
			})
		}

		t.Run("DeleteExtnet", func(t *testing.T) {
			deleteExtnet(ctx, t, profile, extnetNetworkName)
			if t.Failed() && *postMortemLogs {
				PostMortemLogs(t, profile)
			}
		})
	})
}

// connectExtnet create a docker network
func createExtnet(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
	subnet := "172.28.0.0/16"
	ipRange :="172.28.0.0/24"
	cmd := exec.CommandContext(ctx, "docker", "network", "create", extnetNetworkName, fmt.Sprintf("--subnet=%s", subnet), fmt.Sprintf("--ip-range=%s", ipRange))

	result, err := Run(t, cmd)
	if err != nil {
		t.Fatalf("failed to execute 'docker network create', error: %v, output: %s", err, result.Output())
	}
	extnetNetworkID := result.Output()
	fmt.Fprintf(os.Stderr, "%s", extnetNetworkID)
}

// connectExtnet connect network to the minikube cluster
func connectExtnet(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
	cmd := exec.CommandContext(ctx, "docker", "network", "connect", extnetNetworkName, profile)

	result, err := Run(t, cmd)
	if err != nil {
		t.Fatalf("failed to execute 'docker network connect', error: %v, output: %s", err, result.Output())
	}
	extnetNetworkID := result.Output()
	fmt.Fprintf(os.Stderr, "%s", extnetNetworkID)
}

// deleteExtnet just starts a new minikube cluster
func deleteExtnet(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
	cmd := exec.CommandContext(ctx, "docker", "network", "rm", extnetNetworkName)

	result, err := Run(t, cmd)
	if err != nil {
		t.Fatalf("failed to execute 'docker network delete', error: %v, output: %s", err, result.Output())
	}
	extnetNetworkID := result.Output()
	fmt.Fprintf(os.Stderr, "%s", extnetNetworkID)
}

// extnetValidateFreshStart just starts a new minikube cluster
func extnetValidateFreshStart(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
	defer PostMortemLogs(t, profile)

	args := append([]string{"start", "-p", profile, "--memory=2048", "--install-addons=false", "--wait=all"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
}

// extnetValidateStop  runs minikube Stop 
func extnetValidateStop (ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
	defer PostMortemLogs(t, profile)

	args := []string{"stop", "-p", profile, "--alsologtostderr", "-v=5"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to stop minikube with args: %q : %v", rr.Command(), err)
	}
}

// extnetValidateStart runs minikube start
func extnetValidateStart(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
	defer PostMortemLogs(t, profile)

	args := []string{"start", "-p", profile, "--alsologtostderr", "-v=5"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
}

// extnetValidateDelete deletes the cluster
func extnetValidateDelete(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
	defer PostMortemLogs(t, profile)

	args := []string{"delete", "-p", profile, "--alsologtostderr", "-v=5"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to delete minikube with args: %q : %v", rr.Command(), err)
	}
}

// extnetValidateDelete deletes the cluster
func fail(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
	defer PostMortemLogs(t, profile)
  t.Errorf("testing failure")
}

// extnetValidateVerifyDeleted makes sure no left over left after deleting a profile such as containers or volumes
func extnetValidateVerifyDeleted(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
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

// extnetValidateStatus makes sure stopped clusters show up in minikube status correctly
func extnetValidateStatus(ctx context.Context, t *testing.T, profile string, extnetNetworkName string) {
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
