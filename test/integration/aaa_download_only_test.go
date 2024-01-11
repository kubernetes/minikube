//go:build integration

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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// TestDownloadOnly makes sure the --download-only parameter in minikube start caches the appropriate images and tarballs.
func TestDownloadOnly(t *testing.T) { // nolint:gocyclo
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))

	// separate each k8s version testrun into individual profiles to avoid ending up with subsequently mixed up configs like:
	// {Name:download-only-062906 ... KubernetesConfig:{KubernetesVersion:v1.28.4 ...} Nodes:[{Name: IP: Port:8443 KubernetesVersion:v1.16.0 ...}] ...}
	// that will then get artifacts for node's not cluster's KubernetesVersion and fail checks thereafter
	// at the end, cleanup all profiles
	profiles := []string{}
	defer func() {
		for _, profile := range profiles {
			Cleanup(t, profile, cancel)
		}
	}()

	containerRuntime := ContainerRuntime()

	versions := []string{
		constants.OldestKubernetesVersion,
		constants.DefaultKubernetesVersion,
		constants.NewestKubernetesVersion,
	}

	// Small optimization, don't run the exact same set of tests twice
	if constants.DefaultKubernetesVersion == constants.NewestKubernetesVersion {
		versions = versions[:len(versions)-1]
	}

	for _, v := range versions {
		t.Run(v, func(t *testing.T) {
			profile := UniqueProfileName("download-only")
			profiles = append(profiles, profile)
			defer PostMortemLogs(t, profile)

			t.Run("json-events", func(t *testing.T) {
				// --force to avoid uid check
				args := append([]string{"start", "-o=json", "--download-only", "-p", profile, "--force", "--alsologtostderr", fmt.Sprintf("--kubernetes-version=%s", v), fmt.Sprintf("--container-runtime=%s", containerRuntime)}, StartArgs()...)
				rt, err := Run(t, exec.CommandContext(ctx, Target(), args...))
				if err != nil {
					t.Errorf("failed to download only. args: %q %v", args, err)
				}

				s := bufio.NewScanner(bytes.NewReader(rt.Stdout.Bytes()))
				for s.Scan() {
					var rtObj map[string]interface{}
					err := json.Unmarshal(s.Bytes(), &rtObj)
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
				if err := s.Err(); err != nil {
					t.Errorf("failed to read output: %v", err)
				}
			})

			preloadExists := false
			t.Run("preload-exists", func(t *testing.T) {
				// skip for none, as none driver does not have preload feature.
				if NoneDriver() {
					t.Skip("None driver does not have preload")
				}
				// Driver does not matter here, since the only exception is none driver,
				// which cannot occur here.
				if !download.PreloadExists(v, containerRuntime, "docker", true) {
					t.Skip("No preload image")
				}
				// Just make sure the tarball path exists
				if _, err := os.Stat(download.TarballPath(v, containerRuntime)); err != nil {
					t.Errorf("failed to verify preloaded tarball file exists: %v", err)
				}
				preloadExists = true
			})

			t.Run("cached-images", func(t *testing.T) {
				// skip verify for cache images if --driver=none
				if NoneDriver() {
					t.Skip("None driver has no cache")
				}
				if preloadExists {
					t.Skip("Preload exists, images won't be cached")
				}
				imgs, err := images.Kubeadm("", v)
				if err != nil {
					t.Errorf("failed to get kubeadm images for %v: %+v", v, err)
				}

				for _, img := range imgs {
					pathToImage := []string{localpath.MiniPath(), "cache", "images", runtime.GOARCH}
					img = strings.Replace(img, ":", "_", 1) // for example kube-scheduler:v1.15.2 --> kube-scheduler_v1.15.2
					imagePath := strings.Split(img, "/")    // changes "gcr.io/k8s-minikube/storage-provisioner_v5" into ["gcr.io", "k8s-minikube", "storage-provisioner_v5"] to match cache folder structure
					pathToImage = append(pathToImage, imagePath...)
					fp := filepath.Join(pathToImage...)
					_, err := os.Stat(fp)
					if err != nil {
						t.Errorf("expected image file exist at %q but got error: %v", fp, err)
					}
				}
			})

			t.Run("binaries", func(t *testing.T) {
				if preloadExists {
					t.Skip("Preload exists, binaries are present within.")
				}
				// checking binaries downloaded (kubelet,kubeadm)
				for _, bin := range constants.KubernetesReleaseBinaries {
					fp := filepath.Join(localpath.MiniPath(), "cache", "linux", runtime.GOARCH, v, bin)
					_, err := os.Stat(fp)
					if err != nil {
						t.Errorf("expected the file for binary exist at %q but got error %v", fp, err)
					}
				}
			})

			t.Run("kubectl", func(t *testing.T) {
				// If we are on darwin/windows, check to make sure OS specific kubectl has been downloaded
				// as well for the `minikube kubectl` command
				if runtime.GOOS == "linux" {
					t.Skip("Test for darwin and windows")
				}
				binary := "kubectl"
				if runtime.GOOS == "windows" {
					binary = "kubectl.exe"
				}
				fp := filepath.Join(localpath.MiniPath(), "cache", runtime.GOOS, runtime.GOARCH, v, binary)
				if _, err := os.Stat(fp); err != nil {
					t.Errorf("expected the file for binary exist at %q but got error %v", fp, err)
				}
			})

			// checks if the duration of `minikube logs` takes longer than 5 seconds
			t.Run("LogsDuration", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), Seconds(5))
				defer cancel()
				args := []string{"logs", "-p", profile}
				if _, err := Run(t, exec.CommandContext(ctx, Target(), args...)); err != nil {
					t.Logf("minikube logs failed with error: %v", err)
				}
				if err := ctx.Err(); err == context.DeadlineExceeded {
					t.Error("minikube logs expected to finish by 5 seconds, but took longer")
				}
			})

			// This is a weird place to test profile deletion, but this test is serial, and we have a profile to delete!
			t.Run("DeleteAll", func(t *testing.T) {
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

// TestDownloadOnlyKic makes sure --download-only caches the docker driver images as well.
func TestDownloadOnlyKic(t *testing.T) {
	if !KicDriver() {
		t.Skip("skipping, only for docker or podman driver")
	}
	profile := UniqueProfileName("download-docker")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(15))
	defer Cleanup(t, profile, cancel)

	cRuntime := ContainerRuntime()

	args := []string{"start", "--download-only", "-p", profile, "--alsologtostderr"}
	args = append(args, StartArgs()...)

	cmd := exec.CommandContext(ctx, Target(), args...)
	if _, err := Run(t, cmd); err != nil {
		t.Errorf("start with download only failed %q : %v", args, err)
	}

	// Make sure the downloaded image tarball exists
	tarball := download.TarballPath(constants.DefaultKubernetesVersion, cRuntime)
	contents, err := os.ReadFile(tarball)
	if err != nil {
		t.Errorf("failed to read tarball file %q: %v", tarball, err)
	}

	if arm64Platform() {
		t.Skip("Skip for arm64 platform. See https://github.com/kubernetes/minikube/issues/10144")
	}
	// Make sure it has the correct checksum
	checksum := md5.Sum(contents)
	remoteChecksum, err := os.ReadFile(download.PreloadChecksumPath(constants.DefaultKubernetesVersion, cRuntime))
	if err != nil {
		t.Errorf("failed to read checksum file %q : %v", download.PreloadChecksumPath(constants.DefaultKubernetesVersion, cRuntime), err)
	}
	if string(remoteChecksum) != string(checksum[:]) {
		t.Errorf("failed to verify checksum. checksum of %q does not match remote checksum (%q != %q)", tarball, string(remoteChecksum), string(checksum[:]))
	}
}

// createSha256File is a helper function which creates sha256 checksum file from given file
func createSha256File(filePath string) error {
	dat, _ := os.ReadFile(filePath)
	sum := sha256.Sum256(dat)

	f, err := os.Create(filePath + ".sha256")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%x", sum[:]))
	if err != nil {
		return err
	}
	return nil
}

// TestBinaryMirror tests functionality of --binary-mirror flag
func TestBinaryMirror(t *testing.T) {
	profile := UniqueProfileName("binary-mirror")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	defer Cleanup(t, profile, cancel)

	tmpDir := t.TempDir()

	// Start test server which will serve binary files
	ts := httptest.NewServer(
		http.FileServer(http.Dir(tmpDir)),
	)
	defer ts.Close()

	binaryName := "kubectl"
	if runtime.GOOS == "windows" {
		binaryName = "kubectl.exe"
	}
	binaryPath, err := download.Binary(binaryName, constants.DefaultKubernetesVersion, runtime.GOOS, runtime.GOARCH, "")
	if err != nil {
		t.Errorf("Failed to download binary: %+v", err)
	}

	newBinaryDir := filepath.Join(tmpDir, constants.DefaultKubernetesVersion, "bin", runtime.GOOS, runtime.GOARCH)
	if err := os.MkdirAll(newBinaryDir, os.ModePerm); err != nil {
		t.Errorf("Failed to create %s directories", newBinaryDir)
	}

	newBinaryPath := filepath.Join(newBinaryDir, binaryName)
	if err := os.Rename(binaryPath, newBinaryPath); err != nil {
		t.Errorf("Failed to move binary file: %+v", err)
	}
	if err := createSha256File(newBinaryPath); err != nil {
		t.Errorf("Failed to generate sha256 checksum file: %+v", err)
	}

	args := append([]string{"start", "--download-only", "-p", profile, "--alsologtostderr", "--binary-mirror", ts.URL}, StartArgs()...)

	cmd := exec.CommandContext(ctx, Target(), args...)
	if _, err := Run(t, cmd); err != nil {
		t.Errorf("start with --binary-mirror failed %q : %v", args, err)
	}
}
