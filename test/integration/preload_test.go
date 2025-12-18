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
	"os"
	"os/exec"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/localpath"
)

// TestPreload verifies that disabling the initial preload, pulling a specific image, and restarting the cluster preserves the image across restarts.
func TestPreload(t *testing.T) {
	if NoneDriver() {
		t.Skipf("skipping %s - incompatible with none driver", t.Name())
	}

	profile := UniqueProfileName("test-preload")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--memory=3072", "--alsologtostderr", "--wait=true", "--preload=false"}
	startArgs = append(startArgs, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// Now, pull the busybox image into minikube
	image := "gcr.io/k8s-minikube/busybox"
	cmd := exec.CommandContext(ctx, Target(), "-p", profile, "image", "pull", image)
	rr, err = Run(t, cmd)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// stop the cluster
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// re-start the cluster and check if image is preserved with enabled preload
	startArgs = []string{"start", "-p", profile, "--preload=true", "--alsologtostderr", "-v=1", "--wait=true"}
	startArgs = append(startArgs, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
	cmd = exec.CommandContext(ctx, Target(), "-p", profile, "image", "list")
	rr, err = Run(t, cmd)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Output(), image) {
		t.Fatalf("Expected to find %s in image list output, instead got %s", image, rr.Output())
	}
}

// TestPreloadDownloadOnly verifies that downloading preload from github and gcs works
func TestPreloadDownloadOnly(t *testing.T) {
	if NoneDriver() {
		t.Skipf("skipping %s - incompatible with none driver", t.Name())
	}
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

			// Clean up the cache to force download
			cacheDir := localpath.MakeMiniPath("cache", "preloaded-tarball")
			if err := os.RemoveAll(cacheDir); err != nil {
				t.Logf("Failed to clean preload cache at %s: %v", cacheDir, err)
			} else {
				t.Logf("Cleaned preload cache at %s", cacheDir)
			}

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
}
