//go:build integration

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
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
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

// validateMountCmd verifies the minikube mount command works properly
// for the platforms that support it, we're testing:
// - a generic 9p mount
// - a 9p mount on a specific port
// - cleaning-mechanism for profile-specific mounts
func validateMountCmd(ctx context.Context, t *testing.T, profile string) { // nolint
	if NoneDriver() {
		t.Skip("skipping: none driver does not support mount")
	}
	if HyperVDriver() {
		t.Skip("skipping: mount broken on hyperv: https://github.com/kubernetes/minikube/issues/5029")
	}
	if RootlessDriver() {
		t.Skip("skipping: rootless driver does not support mount (because 9p is not mountable inside UserNS)")
	}

	if runtime.GOOS == "windows" {
		t.Skip("skipping: mount broken on windows: https://github.com/kubernetes/minikube/issues/8303")
	}

	t.Run("any-port", func(t *testing.T) {
		tempDir := t.TempDir()

		ctx, cancel := context.WithTimeout(ctx, Minutes(10))

		args := []string{"mount", "-p", profile, fmt.Sprintf("%s:%s", tempDir, guestMount), "--alsologtostderr", "-v=1"}
		ss, err := Start(t, exec.CommandContext(ctx, Target(), args...))
		if err != nil {
			t.Fatalf("%v failed: %v", args, err)
		}

		defer func() {
			if t.Failed() {
				t.Logf("%q failed, getting debug info...", t.Name())
				rr, err := Run(t, exec.Command(Target(), "-p", profile, "ssh", "mount | grep 9p; ls -la /mount-9p; cat /mount-9p/pod-dates"))
				if err != nil {
					t.Logf("debugging command %q failed : %v", rr.Command(), err)
				} else {
					t.Logf("(debug) %s:\n%s", rr.Command(), rr.Stdout)
				}
			}

			// Cleanup in advance of future tests
			rr, err := Run(t, exec.Command(Target(), "-p", profile, "ssh", "sudo umount -f /mount-9p"))
			if err != nil {
				t.Logf("%q: %v", rr.Command(), err)
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
			err := os.WriteFile(p, wantFromTest, 0644)
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
		if err := retry.Expo(checkMount, time.Millisecond*500, Seconds(15)); err != nil {
			// For local testing, allow macOS users to click prompt. If they don't, skip the test.
			if runtime.GOOS == "darwin" {
				t.Skip("skipping: mount did not appear, likely because macOS requires prompt to allow non-code signed binaries to listen on non-localhost port")
			}
			t.Fatalf("/mount-9p did not appear within %s: %v", time.Since(start), err)
		}

		// Assert that we can access the mount without an error. Display for debugging.
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "--", "ls", "-la", guestMount))
		if err != nil {
			t.Fatalf("failed verifying accessing to the mount. args %q : %v", rr.Command(), err)
		}
		t.Logf("guest mount directory contents\n%s", rr.Stdout)

		// Assert that the mount contains our unique test marker, as opposed to a stale mount
		tp := filepath.Join("/mount-9p", testMarker)
		rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "cat", tp))
		if err != nil {
			t.Fatalf("failed to verify the mount contains unique test marked: args %q: %v", rr.Command(), err)
		}

		if !bytes.Equal(rr.Stdout.Bytes(), wantFromTest) {
			// The mount is hosed, exit fast before wasting time launching pods.
			t.Fatalf("%s = %q, want %q", tp, rr.Stdout.Bytes(), wantFromTest)
		}

		// Start the "busybox-mount" pod.
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "busybox-mount-test.yaml")))
		if err != nil {
			t.Fatalf("failed to 'kubectl replace' for busybox-mount-test. args %q : %v", rr.Command(), err)
		}

		if _, err := PodWait(ctx, t, profile, "default", "integration-test=busybox-mount", Minutes(4)); err != nil {
			t.Fatalf("failed waiting for busybox-mount pod: %v", err)
		}

		// Read the file written by pod startup
		p := filepath.Join(tempDir, createdByPod)
		got, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("failed to read file created by pod %q: %v", p, err)
		}
		wantFromPod := []byte("test\n")
		if !bytes.Equal(got, wantFromPod) {
			t.Errorf("the content of the file %q is %q, but want it to be: *%q*", p, got, wantFromPod)
		}

		// test that file written from host was read in by the pod via cat /mount-9p/written-by-host;
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "logs", "busybox-mount"))
		if err != nil {
			t.Errorf("failed to get kubectl logs for busybox-mount. args %q : %v", rr.Command(), err)
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
				t.Errorf("failed to stat the file %q iniside minikube : args %q: %v", gp, rr.Command(), err)
			}

			if runtime.GOOS == "windows" {
				if strings.Contains(rr.Stdout.String(), "Access: 1970-01-01") {
					t.Errorf("expected to get valid access time but got: %q", rr.Stdout)
				}
			}

			if strings.Contains(rr.Stdout.String(), "Modify: 1970-01-01") {
				t.Errorf("expected to get valid modify time but got: %q", rr.Stdout)
			}
		}

		p = filepath.Join(tempDir, createdByTestRemovedByPod)
		if _, err := os.Stat(p); err == nil {
			t.Errorf("expected file %q to be removed but exists !", p)
		}

		p = filepath.Join(tempDir, createdByPodRemovedByTest)
		if err := os.Remove(p); err != nil {
			t.Errorf("failed to remove file %q: %v", p, err)
		}
	})
	t.Run("specific-port", func(t *testing.T) {
		tempDir := t.TempDir()

		ctx, cancel := context.WithTimeout(ctx, Minutes(10))

		args := []string{"mount", "-p", profile, fmt.Sprintf("%s:%s", tempDir, guestMount), "--alsologtostderr", "-v=1", "--port", "46464"}
		ss, err := Start(t, exec.CommandContext(ctx, Target(), args...))
		if err != nil {
			t.Fatalf("%v failed: %v", args, err)
		}

		defer func() {
			if t.Failed() {
				t.Logf("%q failed, getting debug info...", t.Name())
				rr, err := Run(t, exec.Command(Target(), "-p", profile, "ssh", "mount | grep 9p; ls -la /mount-9p; cat /mount-9p/pod-dates"))
				if err != nil {
					t.Logf("debugging command %q failed : %v", rr.Command(), err)
				} else {
					t.Logf("(debug) %s:\n%s", rr.Command(), rr.Stdout)
				}
			}

			// Cleanup in advance of future tests
			rr, err := Run(t, exec.Command(Target(), "-p", profile, "ssh", "sudo umount -f /mount-9p"))
			if err != nil {
				t.Logf("%q: %v", rr.Command(), err)
			}
			ss.Stop(t)
			cancel()
			if *cleanup {
				os.RemoveAll(tempDir)
			}
		}()

		// Block until the mount succeeds to avoid file race
		checkMount := func() error {
			_, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "findmnt -T /mount-9p | grep 9p"))
			return err
		}

		start := time.Now()
		if err := retry.Expo(checkMount, time.Millisecond*500, Seconds(15)); err != nil {
			// For local testing, allow macOS users to click prompt. If they don't, skip the test.
			if runtime.GOOS == "darwin" {
				t.Skip("skipping: mount did not appear, likely because macOS requires prompt to allow non-code signed binaries to listen on non-localhost port")
			}
			t.Fatalf("/mount-9p did not appear within %s: %v", time.Since(start), err)
		}

		// Assert that we can access the mount without an error. Display for debugging.
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "--", "ls", "-la", guestMount))
		if err != nil {
			t.Fatalf("failed verifying accessing to the mount. args %q : %v", rr.Command(), err)
		}
		t.Logf("guest mount directory contents\n%s", rr.Stdout)

		ss.Stop(t)
		t.Logf("reading mount text")
		mountText := func() string {
			str := ""
			var err error
			for err == nil {
				var add string
				add, err = ss.Stdout.ReadString(0)
				str += add
			}
			if err != io.EOF {
				t.Fatalf("failed to read stdout of mount cmd. err: %v", err)
			}
			return str
		}()
		t.Logf("done reading mount text")
		match, err := regexp.Match("Bind Address:\\s*[0-9.]+:46464", []byte(mountText))
		if err != nil {
			t.Fatalf("failed to match regex pattern. err: %v", err)
		}
		if !match {
			t.Fatalf("failed to find bind address with port 46464. Mount command out: \n%v", mountText)
		}
	})

	t.Run("VerifyCleanup", func(t *testing.T) {
		tempDir := t.TempDir()

		ctx, cancel := context.WithTimeout(ctx, Minutes(10))

		guestMountPaths := []string{"/mount1", "/mount2", "/mount3"}

		var mntProcs []*StartSession
		for _, guestMount := range guestMountPaths {
			args := []string{"mount", "-p", profile, fmt.Sprintf("%s:%s", tempDir, guestMount), "--alsologtostderr", "-v=1"}
			mntProc, err := Start(t, exec.CommandContext(ctx, Target(), args...))
			if err != nil {
				t.Fatalf("%v failed: %v", args, err)
			}

			mntProcs = append(mntProcs, mntProc)

		}

		defer func() {
			// Still trying to stop mount processes that could otherwise
			// (if something weird happens...) leave the test run hanging
			// The worst thing that could happen is that we try to kill
			// something that was aleardy killed...
			for _, mp := range mntProcs {
				mp.Stop(t)
			}

			cancel()
			if *cleanup {
				os.RemoveAll(tempDir)
			}
		}()

		// are the mounts alive yet..?
		checkMount := func() error {
			for _, mnt := range guestMountPaths {
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "findmnt -T", mnt))
				if err != nil {
					// if something weird has happened from previous tests..
					// this could at least spare us some waiting
					if strings.Contains(rr.Stdout.String(), fmt.Sprintf("Profile \"%s\" not found.", profile)) {
						t.Fatalf("profile was deleted, cancelling the test")
					}
					return err
				}
			}
			return nil
		}
		if err := retry.Expo(checkMount, time.Millisecond*500, Seconds(15)); err != nil {
			// For local testing, allow macOS users to click prompt. If they don't, skip the test.
			if runtime.GOOS == "darwin" {
				t.Skip("skipping: mount did not appear, likely because macOS requires prompt to allow non-code signed binaries to listen on non-localhost port")
			}
			t.Fatalf("mount was not ready in time: %v", err)
		}

		checkProcsAlive := func(end chan bool) {
			for _, mntp := range mntProcs {
				// Trying to wait for process end
				// if the wait fail with ExitError we know that the process
				// doesn't exist anymore..
				go func(end chan bool) {
					err := mntp.cmd.Wait()
					if _, ok := err.(*exec.ExitError); ok {
						end <- true
					}
				}(end)

				// Either we know that the mount process has ended
				// or we fail after 1 second
				// TODO: is there a better way? rather than waiting..
				select {
				case <-time.After(1 * time.Second):
					t.Fatalf("1s TIMEOUT: Process %d is still running\n", mntp.cmd.Process.Pid)
				case <-end:
					continue
				}
			}
		}

		// exec the mount killer
		_, err := Run(t, exec.Command(Target(), "mount", "-p", profile, "--kill=true"))
		if err != nil {
			t.Fatalf("failed while trying to kill mounts")
		}

		end := make(chan bool, 1)
		checkProcsAlive(end)
	})
}
