// +build integration

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/util/retry"
)

const (
	guestMount                = "/mount-9p"
	createdByPod              = "created-by-pod"
	createdByTest             = "created-by-test"
	createdByTestRemovedByPod = "created-by-test-removed-by-pod"
	createdByPodRemovedByTest = "created-by-pod-removed-by-test"
)

func validateMountCmd(ctx context.Context, t *testing.T, profile string) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support mount")
	}
	if HyperVDriver() {
		t.Skip("skipping: mount broken on hyperv: https://github.com/kubernetes/minikube/issues/5029")
	}

	tempDir, err := ioutil.TempDir("", "mounttest")
	if err != nil {
		t.Fatalf("Unexpected error while creating tempDir: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)

	args := []string{"mount", "-p", profile, fmt.Sprintf("%s:%s", tempDir, guestMount), "--alsologtostderr", "-v=1"}
	ss, err := Start(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("%v failed: %v", args, err)
	}

	defer func() {
		if t.Failed() {
			t.Logf("%s failed, getting debug info...", t.Name())
			rr, err := Run(t, exec.Command(Target(), "-p", profile, "ssh", "mount | grep 9p; ls -la /mount-9p; cat /mount-9p/pod-dates"))
			if err != nil {
				t.Logf("%s: %v", rr.Command(), err)
			} else {
				t.Logf("(debug) %s:\n%s", rr.Command(), rr.Stdout)
			}
		}

		// Cleanup in advance of future tests
		rr, err := Run(t, exec.Command(Target(), "-p", profile, "ssh", "sudo umount -f /mount-9p"))
		if err != nil {
			t.Logf("%s: %v", rr.Command(), err)
		}
		ss.Stop(t)
		cancel()
		if *cleanup {
			os.RemoveAll(tempDir)
		}
	}()

	// Write local files
	testMarker := fmt.Sprintf("test-%d", time.Now().UnixNano())
	wantFromTest := []byte(testMarker)
	for _, name := range []string{createdByTest, createdByTestRemovedByPod, testMarker} {
		p := filepath.Join(tempDir, name)
		err := ioutil.WriteFile(p, wantFromTest, 0644)
		t.Logf("wrote %q to %s", wantFromTest, p)
		if err != nil {
			t.Errorf("WriteFile %s: %v", p, err)
		}
	}

	// Block until the mount succeeds to avoid file race
	checkMount := func() error {
		_, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "findmnt -T /mount-9p | grep 9p"))
		return err
	}

	start := time.Now()
	if err := retry.Expo(checkMount, time.Second, 15*time.Second); err != nil {
		// For local testing, allow macOS users to click prompt. If they don't, skip the test.
		if runtime.GOOS == "darwin" {
			t.Skip("skipping: mount did not appear, likely because macOS requires prompt to allow non-codesigned binaries to listen on non-localhost port")
		}
		t.Fatalf("/mount-9p did not appear within %s: %v", time.Since(start), err)
	}

	// Assert that we can access the mount without an error. Display for debugging.
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "--", "ls", "-la", guestMount))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	t.Logf("guest mount directory contents\n%s", rr.Stdout)

	// Assert that the mount contains our unique test marker, as opposed to a stale mount
	tp := filepath.Join("/mount-9p", testMarker)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "cat", tp))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	if !bytes.Equal(rr.Stdout.Bytes(), wantFromTest) {
		// The mount is hosed, exit fast before wasting time launching pods.
		t.Fatalf("%s = %q, want %q", tp, rr.Stdout.Bytes(), wantFromTest)
	}

	// Start the "busybox-mount" pod.
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "busybox-mount-test.yaml")))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "integration-test=busybox-mount", 4*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}

	// Read the file written by pod startup
	p := filepath.Join(tempDir, createdByPod)
	got, err := ioutil.ReadFile(p)
	if err != nil {
		t.Errorf("readfile %s: %v", p, err)
	}
	wantFromPod := []byte("test\n")
	if !bytes.Equal(got, wantFromPod) {
		t.Errorf("%s = %q, want %q", p, got, wantFromPod)
	}

	// test that file written from host was read in by the pod via cat /mount-9p/written-by-host;
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "logs", "busybox-mount"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if !bytes.Equal(rr.Stdout.Bytes(), wantFromTest) {
		t.Errorf("busybox-mount logs = %q, want %q", rr.Stdout.Bytes(), wantFromTest)
	}

	// test file timestamps are correct
	for _, name := range []string{createdByTest, createdByPod} {
		gp := path.Join(guestMount, name)
		// test that file written from host was read in by the pod via cat /mount-9p/fromhost;
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "stat", gp))
		if err != nil {
			t.Errorf("%s failed: %v", rr.Args, err)
		}

		if runtime.GOOS == "windows" {
			if strings.Contains(rr.Stdout.String(), "Access: 1970-01-01") {
				t.Errorf("invalid access time: %v", rr.Stdout)
			}
		}

		if strings.Contains(rr.Stdout.String(), "Modify: 1970-01-01") {
			t.Errorf("invalid modify time: %v", rr.Stdout)
		}
	}

	p = filepath.Join(tempDir, createdByTestRemovedByPod)
	if _, err := os.Stat(p); err == nil {
		t.Errorf("expected file %s to be removed", p)
	}

	p = filepath.Join(tempDir, createdByPodRemovedByTest)
	if err := os.Remove(p); err != nil {
		t.Errorf("unexpected error removing file %s: %v", p, err)
	}
}
