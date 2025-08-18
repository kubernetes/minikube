//go:build integration

/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/test/integration/findmnt"
)

const (
	mountGID   = "0"
	mountMSize = "6543"
	mountUID   = "0"
	guestPath  = "/minikube-host"
)

var mountStartPort = 46463

func mountPort() string {
	return strconv.Itoa(mountStartPort)
}

// TestMountStart tests using the mount command on start
func TestMountStart(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support mount")
	}

	type validateFunc func(context.Context, *testing.T, string)
	profile1 := UniqueProfileName("mount-start-1")
	profile2 := UniqueProfileName("mount-start-2")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(15))
	defer Cleanup(t, profile1, cancel)
	defer Cleanup(t, profile2, cancel)

	// Serial tests
	t.Run("serial", func(t *testing.T) {
		hostPath := t.TempDir()
		startProfileWithMount := func(ctx context.Context, t *testing.T, profile string) {
			validateStartWithMount(ctx, t, profile, hostPath)
		}

		tests := []struct {
			name      string
			validator validateFunc
			profile   string
		}{
			{"StartWithMountFirst", startProfileWithMount, profile1},
			{"VerifyMountFirst", validateMount, profile1},
			{"StartWithMountSecond", startProfileWithMount, profile2},
			{"VerifyMountSecond", validateMount, profile2},
			{"DeleteFirst", validateDelete, profile1},
			{"VerifyMountPostDelete", validateMount, profile2},
			{"Stop", validateMountStop, profile2},
			{"RestartStopped", validateRestart, profile2},
			{"VerifyMountPostStop", validateMount, profile2},
		}

		for _, test := range tests {
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			if t.Failed() {
				t.Fatalf("Previous test failed, not running dependent tests")
			}

			t.Run(test.name, func(t *testing.T) {
				test.validator(ctx, t, test.profile)
			})
		}
	})
}

// validateStartWithMount starts a cluster with mount enabled
func validateStartWithMount(ctx context.Context, t *testing.T, profile string, hostPath string) {
	defer PostMortemLogs(t, profile)

	// We have to increment this because if you have two mounts with the same port, when you kill one cluster the mount will break for the other
	mountStartPort++

	args := []string{
		"start",
		"-p", profile,
		"--memory=3072",
		"--mount-string", fmt.Sprintf("%s:%s", hostPath, guestPath),
		"--mount-gid", mountGID,
		"--mount-msize", mountMSize,
		"--mount-port", mountPort(),
		"--mount-uid", mountUID,
		"--no-kubernetes",
	}
	args = append(args, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
	// The mount takes a split second to come up, without this the validateMount test will fail
	time.Sleep(1 * time.Second)
}

// validateMount checks if the cluster has a folder mounted
func validateMount(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	sshArgs := []string{"-p", profile, "ssh", "--"}

	args := sshArgs
	args = append(args, "ls", guestPath)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("mount failed: %q : %v", rr.Command(), err)
	}

	// Docker has it's own mounting method, it doesn't respect the mounting flags
	// We can't get the mount details with Hyper-V
	if DockerDriver() || HyperVDriver() {
		return
	}

	args = sshArgs
	args = append(args, "findmnt", "--json", guestPath)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("command failed %q: %v", rr.Command(), err)
	}

	result, err := findmnt.ParseOutput(rr.Stdout.Bytes())
	if err != nil {
		t.Fatalf("failed to parse output %s: %v", rr.Stdout.String(), err)
	}

	if len(result.Filesystems) == 0 {
		t.Fatalf("no filesystems found")
	}

	fs := result.Filesystems[0]
	if fs.FSType == constants.FSTypeVirtiofs {
		t.Logf("Options not supported yet for %q filesystem", fs.FSType)
		return
	}

	if fs.FSType != constants.FSType9p {
		t.Fatalf("unexpected filesystem %q", fs.FSType)
	}

	options := strings.Split(fs.Options, ",")

	flags := []struct {
		key      string
		expected string
	}{
		{"dfltgid", mountGID},
		{"dfltuid", mountUID},
		{"msize", mountMSize},
		{"port", mountPort()},
	}

	for _, flag := range flags {
		want := fmt.Sprintf("%s=%s", flag.key, flag.expected)
		if !slices.Contains(options, want) {
			t.Errorf("wanted %s to be: %q; got: %q", flag.key, want, options)
		}
	}
}

// validateMountStop stops a cluster
func validateMountStop(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := []string{"stop", "-p", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("stop failed: %q : %v", rr.Command(), err)
	}
}

// validateRestart restarts a cluster
func validateRestart(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := []string{"start", "-p", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("restart failed: %q : %v", rr.Command(), err)
	}
	// The mount takes a split second to come up, without this the validateMount test will fail
	time.Sleep(1 * time.Second)
}
