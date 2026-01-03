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
// also tests --preload-source should work for both github and gcs
func TestPreload(t *testing.T) {
	MaybeParallel(t)
	if NoneDriver() {
		t.Skipf("skipping %s - incompatible with none driver", t.Name())
	}

	if ContainerRuntime() == "crio" {
		t.Skipf("skipping %s - user-pulled images not persisted across restarts with crio", t.Name())
	}

	profile := UniqueProfileName("test-preload")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer CleanupWithLogs(t, profile, cancel)

	userImage := busyboxImage

	// These subtests run sequentially (t.Run blocks until completion) to share the same profile/cluster state.
	if t.Run("Start-NoPreload-PullImage", func(t *testing.T) {
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
	}) {
		t.Run("Restart-With-Preload-Check-User-Image", func(t *testing.T) {
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
	}

	// PreloadSrc verifies that downloading preload from github and gcs works using --preload-src and --download-only
	// "auto" is the default preload source (tries both gcs and github); here we explicitly verify each source
	t.Run("PreloadSrc", func(t *testing.T) {
		MaybeParallel(t)
		tests := []struct {
			name              string
			source            string
			kubernetesVersion string // using versions that are not used in the test to make sure they dont pre-exist
			wantLog           string
		}{
			{"gcs", "gcs", "v1.34.0-rc.1", "Downloading preload from https://storage.googleapis.com"},
			{"github", "github", "v1.34.0-rc.2", "Downloading preload from https://github.com"},
			{"gcs-cached", "gcs", "v1.34.0-rc.2", "in cache, skipping download"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				profile := UniqueProfileName("test-preload-dl-" + tc.name)
				ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
				defer CleanupWithLogs(t, profile, cancel)

				startArgs := []string{"start", "-p", profile, "--download-only", "--kubernetes-version", tc.kubernetesVersion, fmt.Sprintf("--preload-source=%s", tc.source), "--alsologtostderr", "--v=1"}
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
