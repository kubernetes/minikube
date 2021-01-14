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
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestDownloadOnly(t *testing.T) {
	for _, r := range []string{"crio", "docker", "containerd"} {
		t.Run(r, func(t *testing.T) {
			// Stores the startup run result for later error messages
			var rrr *RunResult

			profile := UniqueProfileName(r)
			ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
			defer Cleanup(t, profile, cancel)

			versions := []string{
				constants.OldestKubernetesVersion,
				constants.DefaultKubernetesVersion,
				constants.NewestKubernetesVersion,
			}

			for _, v := range versions {
				t.Run(v, func(t *testing.T) {
					defer PostMortemLogs(t, profile)

					// --force to avoid uid check
					args := append([]string{"start", "-o=json", "--download-only", "-p", profile, "--force", "--alsologtostderr", fmt.Sprintf("--kubernetes-version=%s", v), fmt.Sprintf("--container-runtime=%s", r)}, StartArgs()...)

					rt, err := Run(t, exec.CommandContext(ctx, Target(), args...))
					if rrr == nil {
						// Preserve the initial run-result for debugging
						rrr = rt
					}
					if err != nil {
						t.Errorf("failed to download only. args: %q %v", args, err)
					}
					t.Run("check json events", func(t *testing.T) {
						s := bufio.NewScanner(bytes.NewReader(rt.Stdout.Bytes()))
						for s.Scan() {
							var rtObj map[string]interface{}
							err = json.Unmarshal(s.Bytes(), &rtObj)
							if err != nil {
								t.Errorf("failed to parse output: %v", err)
							} else if step, ok := rtObj["data"]; ok {
								if stepMap, ok := step.(map[string]interface{}); ok {
									if stepMap["currentstep"] == "" {
										t.Errorf("Empty step number for %v", stepMap["name"])
									}
								}
							}
						}
					})

					// skip for none, as none driver does not have preload feature.
					if !NoneDriver() {
						if download.PreloadExists(v, r, true) {
							// Just make sure the tarball path exists
							if _, err := os.Stat(download.TarballPath(v, r)); err != nil {
								t.Errorf("failed to verify preloaded tarball file exists: %v", err)
							}
							return
						}
					}
					imgs, err := images.Kubeadm("", v)
					if err != nil {
						t.Errorf("failed to get kubeadm images for %v: %+v", v, err)
					}

					// skip verify for cache images if --driver=none
					if !NoneDriver() {
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
					for _, bin := range constants.KubernetesReleaseBinaries {
						fp := filepath.Join(localpath.MiniPath(), "cache", "linux", v, bin)
						_, err := os.Stat(fp)
						if err != nil {
							t.Errorf("expected the file for binary exist at %q but got error %v", fp, err)
						}
					}

					// If we are on darwin/windows, check to make sure OS specific kubectl has been downloaded
					// as well for the `minikube kubectl` command
					if runtime.GOOS == "linux" {
						return
					}
					binary := "kubectl"
					if runtime.GOOS == "windows" {
						binary = "kubectl.exe"
					}
					fp := filepath.Join(localpath.MiniPath(), "cache", runtime.GOOS, v, binary)
					if _, err := os.Stat(fp); err != nil {
						t.Errorf("expected the file for binary exist at %q but got error %v", fp, err)
					}
				})
			}

			// This is a weird place to test profile deletion, but this test is serial, and we have a profile to delete!
			t.Run("DeleteAll", func(t *testing.T) {
				defer PostMortemLogs(t, profile)

				if !CanCleanup() {
					t.Skip("skipping, as cleanup is disabled")
				}
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "delete", "--all"))
				if err != nil {
					t.Errorf("failed to delete all. args: %q : %v", rr.Command(), err)
				}
			})
			// Delete should always succeed, even if previously partially or fully deleted.
			t.Run("DeleteAlwaysSucceeds", func(t *testing.T) {
				defer PostMortemLogs(t, profile)

				if !CanCleanup() {
					t.Skip("skipping, as cleanup is disabled")
				}
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "delete", "-p", profile))
				if err != nil {
					t.Errorf("failed to delete. args: %q: %v", rr.Command(), err)
				}
			})
		})
	}
}

func TestDownloadOnlyKic(t *testing.T) {
	if !KicDriver() {
		t.Skip("skipping, only for docker or podman driver")
	}
	profile := UniqueProfileName("download-docker")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(15))
	defer Cleanup(t, profile, cancel)

	// TODO: #7795 add containerd to download only too
	cRuntime := "docker"

	args := []string{"start", "--download-only", "-p", profile, "--force", "--alsologtostderr"}
	args = append(args, StartArgs()...)

	cmd := exec.CommandContext(ctx, Target(), args...)
	// make sure this works even if docker daemon isn't running
	cmd.Env = append(os.Environ(), "DOCKER_HOST=/does/not/exist")
	if _, err := Run(t, cmd); err != nil {
		t.Errorf("start with download only failed %q : %v", args, err)
	}

	// Make sure the downloaded image tarball exists
	tarball := download.TarballPath(constants.DefaultKubernetesVersion, cRuntime)
	contents, err := ioutil.ReadFile(tarball)
	if err != nil {
		t.Errorf("failed to read tarball file %q: %v", tarball, err)
	}
	if !arm64Platform() {
		// Make sure it has the correct checksum
		checksum := md5.Sum(contents)
		remoteChecksum, err := ioutil.ReadFile(download.PreloadChecksumPath(constants.DefaultKubernetesVersion, cRuntime))
		if err != nil {
			t.Errorf("failed to read checksum file %q : %v", download.PreloadChecksumPath(constants.DefaultKubernetesVersion, cRuntime), err)
		}
		if string(remoteChecksum) != string(checksum[:]) {
			t.Errorf("failed to verify checksum. checksum of %q does not match remote checksum (%q != %q)", tarball, string(remoteChecksum), string(checksum[:]))
		}
	}
}
