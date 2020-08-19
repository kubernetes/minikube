// +build integration

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"os"
	"os/exec"
	"path"
	"testing"

	"k8s.io/minikube/cmd/minikube/cmd"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestInsufficientStorage(t *testing.T) {
	if !KicDriver() {
		t.Skip("only runs with docker driver")
	}
	profile := UniqueProfileName("insufficient-storage")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer Cleanup(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--output=json", "--wait=true"}
	startArgs = append(startArgs, StartArgs()...)
	c := exec.CommandContext(ctx, Target(), startArgs...)
	// artificially set /var to 100% capacity
	c.Env = append(os.Environ(), fmt.Sprintf("%s=100", constants.TestMemoryCapacityEnv))

	rr, err := Run(t, c)
	if err == nil {
		t.Fatalf("expected command to fail, but it succeeded: %v\n%v", rr.Command(), err)
	}
	stdout := runStatusCmd(t, ctx, profile)
	verifyClusterState(t, stdout)

	// try deleting events.json and make sure this still works
	eventsFile := path.Join(localpath.MiniPath(), "profiles", profile, "events.json")
	if err := os.Remove(eventsFile); err != nil {
		t.Fatalf("removing %s", eventsFile)
	}
	stdout = runStatusCmd(t, ctx, profile)
	verifyClusterState(t, stdout)
}

// runStatusCmd runs the status command and returns stdout
func runStatusCmd(t *testing.T, ctx context.Context, profile string) []byte {
	// make sure minikube status shows insufficient storage
	c := exec.CommandContext(ctx, Target(), "status", "-p", profile, "--output=json", "--layout=cluster")
	// artificially set /var to 100% capacity
	c.Env = append(os.Environ(), fmt.Sprintf("%s=100", constants.TestMemoryCapacityEnv))
	rr, err := Run(t, c)
	// status exits non-0 if status isn't Running
	if err == nil {
		t.Fatalf("expected command to fail, but it succeeded: %v\n%v", rr.Command(), err)
	}
	return rr.Stdout.Bytes()
}

func verifyClusterState(t *testing.T, contents []byte) {
	var cs cmd.ClusterState
	if err := json.Unmarshal(contents, &cs); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	// verify the status looks as we expect
	if cs.StatusCode != cmd.InsufficientStorage {
		t.Fatalf("incorrect status code: %v", cs.StatusCode)
	}
	if cs.StatusName != "InsufficientStorage" {
		t.Fatalf("incorrect status name: %v", cs.StatusName)
	}
	for _, n := range cs.Nodes {
		if n.StatusCode != cmd.InsufficientStorage {
			t.Fatalf("incorrect node status code: %v", cs.StatusCode)
		}
	}
}
