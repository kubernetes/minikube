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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestDownloadAndDeleteAll(t *testing.T) {
	profile := UniqueProfileName("download")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer Cleanup(t, profile, cancel)

	t.Run("group", func(t *testing.T) {
		versions := []string{
			constants.OldestKubernetesVersion,
			constants.DefaultKubernetesVersion,
			constants.NewestKubernetesVersion,
		}
		for _, v := range versions {
			t.Run(v, func(t *testing.T) {
				args := append([]string{"start", "--download-only", "-p", profile, fmt.Sprintf("--kubernetes-version=%s", v)}, StartArgs()...)
				_, err := Run(t, exec.CommandContext(ctx, Target(), args...))
				if err != nil {
					t.Errorf("%s failed: %v", args, err)
				}

				// None driver does not cache images, so this test will fail
				if !NoneDriver() {
					imgs := images.CachedImages("", v)
					for _, img := range imgs {
						img = strings.Replace(img, ":", "_", 1) // for example kube-scheduler:v1.15.2 --> kube-scheduler_v1.15.2
						fp := filepath.Join(localpath.MiniPath(), "cache", "images", img)
						_, err := os.Stat(fp)
						if err != nil {
							t.Errorf("expected image file exist at %q but got error: %v", fp, err)
						}
					}
				}

				// checking binaries downloaded (kubelet,kubeadm)
				for _, bin := range constants.KubeadmBinaries {
					fp := filepath.Join(localpath.MiniPath(), "cache", v, bin)
					_, err := os.Stat(fp)
					if err != nil {
						t.Errorf("expected the file for binary exist at %q but got error %v", fp, err)
					}
				}
			})
		}
		// This is a weird place to test profile deletion, but this test is serial, and we have a profile to delete!
		t.Run("DeleteAll", func(t *testing.T) {
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "delete", "--all"))
			if err != nil {
				t.Errorf("%s failed: %v", rr.Args, err)
			}
		})
		// Delete should always succeed, even if previously partially or fully deleted.
		t.Run("DeleteAlwaysSucceeds", func(t *testing.T) {
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "delete", "-p", profile))
			if err != nil {
				t.Errorf("%s failed: %v", rr.Args, err)
			}
		})
	})

}
