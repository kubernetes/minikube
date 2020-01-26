// +build integration

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestDownloadOnly(t *testing.T) {
	profile := UniqueProfileName("download")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer Cleanup(t, profile, cancel)

	// Stores the startup run result for later error messages
	var rrr *RunResult
	var err error

	t.Run("group", func(t *testing.T) {
		versions := []string{
			constants.OldestKubernetesVersion,
			constants.DefaultKubernetesVersion,
			constants.NewestKubernetesVersion,
		}
		for _, v := range versions {
			t.Run(v, func(t *testing.T) {
				// Explicitly does not pass StartArgs() to test driver default
				// --force to avoid uid check
				args := []string{"start", "--download-only", "-p", profile, "--force", "--alsologtostderr", fmt.Sprintf("--kubernetes-version=%s", v)}

				// Preserve the initial run-result for debugging
				if rrr == nil {
					rrr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
				} else {
					_, err = Run(t, exec.CommandContext(ctx, Target(), args...))
				}

				if err != nil {
					t.Errorf("%s failed: %v", args, err)
				}

				imgs, err := images.Kubeadm("", v)
				if err != nil {
					t.Errorf("kubeadm images: %v", v)
				}

				for _, img := range imgs {
					img = strings.Replace(img, ":", "_", 1) // for example kube-scheduler:v1.15.2 --> kube-scheduler_v1.15.2
					fp := filepath.Join(localpath.MiniPath(), "cache", "images", img)
					_, err := os.Stat(fp)
					if err != nil {
						t.Errorf("expected image file exist at %q but got error: %v", fp, err)
					}
				}

				// checking binaries downloaded (kubelet,kubeadm)
				for _, bin := range constants.KubernetesReleaseBinaries {
					fp := filepath.Join(localpath.MiniPath(), "cache", v, bin)
					_, err := os.Stat(fp)
					if err != nil {
						t.Errorf("expected the file for binary exist at %q but got error %v", fp, err)
					}
				}
			})
		}

		// Check that the profile we've created has the expected driver
		t.Run("ExpectedDefaultDriver", func(t *testing.T) {
			if ExpectedDefaultDriver() == "" {
				t.Skipf("--expected-default-driver is unset, skipping test")
				return
			}
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "--output", "json"))
			if err != nil {
				t.Errorf("%s failed: %v", rr.Args, err)
			}
			var ps map[string][]config.Profile
			err = json.Unmarshal(rr.Stdout.Bytes(), &ps)
			if err != nil {
				t.Errorf("%s failed: %v", rr.Args, err)
			}

			got := ""
			for _, p := range ps["valid"] {
				if p.Name == profile {
					got = p.Config.VMDriver
				}
			}

			if got != ExpectedDefaultDriver() {
				t.Errorf("got driver %q, expected %q\nstart output: %s", got, ExpectedDefaultDriver(), rrr.Output())
			}
		})

		// This is a weird place to test profile deletion, but this test is serial, and we have a profile to delete!
		t.Run("DeleteAll", func(t *testing.T) {
			if !CanCleanup() {
				t.Skip("skipping, as cleanup is disabled")
			}
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "delete", "--all"))
			if err != nil {
				t.Errorf("%s failed: %v", rr.Args, err)
			}
		})
		// Delete should always succeed, even if previously partially or fully deleted.
		t.Run("DeleteAlwaysSucceeds", func(t *testing.T) {
			if !CanCleanup() {
				t.Skip("skipping, as cleanup is disabled")
			}
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "delete", "-p", profile))
			if err != nil {
				t.Errorf("%s failed: %v", rr.Args, err)
			}
		})
	})

}
