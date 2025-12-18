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
	"os/exec"
	"strings"
	"testing"
)

// TestPreload verifies that disabling the initial preload, pulling a specific image,
// and restarting the cluster preserves the image across restarts.
// also tests --preload-src should work for both github and gcs
func TestPreload(t *testing.T) {
	MaybeParallel(t)
	if NoneDriver() {
		t.Skipf("skipping %s - incompatible with none driver", t.Name())
	}

	profile := UniqueProfileName("test-preload")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer CleanupWithLogs(t, profile, cancel)

	userImage := "public.ecr.aws/docker/library/busybox:latest"

	// These subtests run sequentially (t.Run blocks until completion) to share the same profile/cluster state.
	t.Run("StartNoPreloadAndPullImage", func(t *testing.T) {
		startArgs := []string{"start", "-p", profile, "--memory=3072", "--alsologtostderr", "--wait=true", "--preload=false"}
		startArgs = append(startArgs, StartArgs()...)

		rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}

		// Now, pull the busybox image into minikube
		cmd := exec.CommandContext(ctx, Target(), "-p", profile, "image", "pull", userImage)
		rr, err = Run(t, cmd)
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}

		// stop the cluster
		rr, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}
	})

	t.Run("RestartWithPreloadAndCheckUserImage", func(t *testing.T) {
		// re-start the cluster and check if image is preserved with enabled preload
		startArgs := []string{"start", "-p", profile, "--preload=true", "--alsologtostderr", "-v=1", "--wait=true"}
		startArgs = append(startArgs, StartArgs()...)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}
		cmd := exec.CommandContext(ctx, Target(), "-p", profile, "image", "list")
		rr, err = Run(t, cmd)
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}
		if !strings.Contains(rr.Output(), userImage) {
			t.Fatalf("Expected to find %s in image list output, instead got %s", userImage, rr.Output())
		}
	})
	// PreloadSrc verifies that downloading preload from github and gcs works using --preload-src and --download-only
	t.Run("PreloadSrc", func(t *testing.T) {
		MaybeParallel(t)
		// "gcs" is the default source, so we can verify it by default
		tests := []struct {
			name    string
			source  string
			wantLog string
		}{
			{"gcs", "gcs", "Downloading preload from https://storage.googleapis.com"},
			{"github", "github", "Downloading preload from https://github.com"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				profile := UniqueProfileName("test-preload-dl-" + tc.name)
				ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
				defer CleanupWithLogs(t, profile, cancel)

				startArgs := []string{"start", "-p", profile, "--download-only", fmt.Sprintf("--preload-src=%s", tc.source), "--alsologtostderr", "--v=1"}
				startArgs = append(startArgs, StartArgs()...)

				rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
				if err != nil {
					t.Fatalf("%s failed: %v", rr.Command(), err)
				}
				if !strings.Contains(rr.Output(), tc.wantLog) {
					t.Fatalf("Expected to find %q in output, but got:\n%s", tc.wantLog, rr.Output())
				}
			})
		}
	})
}
