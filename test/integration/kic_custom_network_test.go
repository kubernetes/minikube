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
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/drivers/kic/oci"
)

// TestKicCustomNetwork verifies the docker driver works with a custom network
func TestKicCustomNetwork(t *testing.T) {
	if !KicDriver() {
		t.Skip("only runs with docker driver")
	}

	tests := []struct {
		description string
		networkName string
	}{
		{
			description: "create custom network",
		}, {
			description: "use default bridge network",
			networkName: "bridge",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			profile := UniqueProfileName("docker-network")
			ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
			defer Cleanup(t, profile, cancel)

			startArgs := []string{"start", "-p", profile, fmt.Sprintf("--network=%s", test.networkName)}
			c := exec.CommandContext(ctx, Target(), startArgs...)
			rr, err := Run(t, c)
			if err != nil {
				t.Fatalf("%v failed: %v\n%v", rr.Command(), err, rr.Output())
			}
			nn := test.networkName
			if nn == "" {
				nn = profile
			}
			verifyNetworkExists(ctx, t, nn)
		})
	}
}

// TestKicExistingNetwork verifies the docker driver and run with an existing network
func TestKicExistingNetwork(t *testing.T) {
	if !KicDriver() {
		t.Skip("only runs with docker driver")
	}
	// create custom network
	networkName := "existing-network"
	if _, err := oci.CreateNetwork(oci.Docker, networkName); err != nil {
		t.Fatalf("error creating network: %v", err)
	}
	defer func() {
		if err := oci.DeleteKICNetworks(oci.Docker); err != nil {
			t.Logf("error deleting kic network, may need to delete manually: %v", err)
		}
	}()
	profile := UniqueProfileName("existing-network")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer Cleanup(t, profile, cancel)

	verifyNetworkExists(ctx, t, networkName)

	startArgs := []string{"start", "-p", profile, fmt.Sprintf("--network=%s", networkName)}
	c := exec.CommandContext(ctx, Target(), startArgs...)
	rr, err := Run(t, c)
	if err != nil {
		t.Fatalf("%v failed: %v\n%v", rr.Command(), err, rr.Output())
	}
}

func verifyNetworkExists(ctx context.Context, t *testing.T, networkName string) {
	c := exec.CommandContext(ctx, "docker", "network", "ls", "--format", "{{.Name}}")
	rr, err := Run(t, c)
	if err != nil {
		t.Fatalf("%v failed: %v\n%v", rr.Command(), err, rr.Output())
	}
	if output := rr.Output(); !strings.Contains(output, networkName) {
		t.Fatalf("%s network is not listed by [%v]: %v", networkName, c.Args, output)
	}
}
