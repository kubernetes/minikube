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
	"path/filepath"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
)

// uncachedPreloadVersions returns up to count stable, supported Kubernetes versions that are not in the given
// minikube home’s preload cache, starting with the oldest version.
func uncachedPreloadVersions(t *testing.T, minikubeHome, containerRuntime string, count int) []string {
	t.Helper()
	oldest := semver.MustParse(strings.TrimPrefix(constants.OldestKubernetesVersion, "v"))
	newest := semver.MustParse(strings.TrimPrefix(constants.NewestKubernetesVersion, "v"))
	versions := make([]string, 0, count)
	for i := len(constants.ValidKubernetesVersions) - 1; i >= 0; i-- {
		k8sVersion := constants.ValidKubernetesVersions[i]
		// ParseTolerant normalizes historical shortened entries like v0.4 so the supported-range check can skip them.
		parsed, err := semver.ParseTolerant(k8sVersion)
		if err != nil {
			t.Fatalf("invalid Kubernetes version %q: %v", k8sVersion, err)
		}
		if len(parsed.Pre) != 0 || parsed.LT(oldest) || parsed.GT(newest) {
			continue
		}
		cacheDir := filepath.Join(minikubeHome, "cache", "preloaded-tarball")
		tarballPath := filepath.Join(cacheDir, download.TarballName(k8sVersion, containerRuntime))
		info, err := os.Stat(tarballPath)
		if err == nil {
			if info.Size() > 0 {
				continue
			}
		} else if !os.IsNotExist(err) {
			t.Fatalf("failed to inspect preload cache for %s: %v", k8sVersion, err)
		}
		versions = append(versions, k8sVersion)
		if len(versions) == count {
			return versions
		}
	}
	return versions
}

// TestPreload verifies that disabling the initial preload, pulling a specific image,
// and restarting the cluster preserves the image across restarts.
// also tests --preload-source should work for both github and gcs
func TestPreload(t *testing.T) {
	MaybeParallel(t)
	if NoneDriver() {
		t.Skipf("skipping %s - incompatible with none driver", t.Name())
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
			// containerd preload overwrites /var and drops user images; unlike docker
			// and cri-o it has no backup/restore yet (#22269), so this check can't pass.
			if ContainerRuntime() == constants.Containerd {
				t.Skip("containerd preload drops user images, see https://github.com/kubernetes/minikube/issues/23035")
			}
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

	// PreloadSrc verifies that downloading preload from GitHub and GCS works using --preload-source and --download-only.
	// "auto" is the default preload source (tries both gcs and github); here we explicitly verify each source
	t.Run("PreloadSrc", func(t *testing.T) {
		MaybeParallel(t)
		// Use a private cache so other tests and the developer's existing cache cannot affect these source checks.
		minikubeHome := filepath.Join(t.TempDir(), ".minikube")
		versions := uncachedPreloadVersions(t, minikubeHome, ContainerRuntime(), 2)
		if len(versions) < 2 {
			t.Skipf("need two uncached supported preload versions to test download sources, found %d", len(versions))
		}
		tests := []struct {
			name              string
			source            string
			kubernetesVersion string
			wantLog           string
		}{
			{"gcs", "gcs", versions[0], "Downloading preload from https://storage.googleapis.com"},
			{"github", "github", versions[1], "Downloading preload from https://github.com"},
			{"gcs-cached", "gcs", versions[1], "in cache, skipping download"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.name == "gcs-cached" && DockerDriver() {
					t.Skip("skipping: flaky on Docker driver: https://github.com/kubernetes/minikube/issues/23164")
				}
				profile := UniqueProfileName("test-preload-dl-" + tc.name)
				ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
				defer cancel()

				startArgs := []string{"start", "-p", profile, "--download-only", "--kubernetes-version", tc.kubernetesVersion, fmt.Sprintf("--preload-source=%s", tc.source), "--alsologtostderr", "--v=1"}
				startArgs = append(startArgs, StartArgs()...)

				cmd := exec.CommandContext(ctx, Target(), startArgs...)
				cmd.Env = append(os.Environ(), "MINIKUBE_HOME="+minikubeHome)
				rr, err := Run(t, cmd)
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
