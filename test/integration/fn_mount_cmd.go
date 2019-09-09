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
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const guestMount = "/mount-9p"

func validateMountCmd(ctx context.Context, t *testing.T, profile string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if NoneDriver() {
		t.Skip("skipping test for none driver as it does not need mount")
	}

	tempDir, err := ioutil.TempDir("", "mounttest")
	if err != nil {
		t.Fatalf("Unexpected error while creating tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Start the mount (NOTE: add -v=1 for protocol level dumps)
	args := []string{"mount", "-p", profile, fmt.Sprintf("%s:%s", tempDir, guestMount), "--alsologtostderr"}
	ss, err := StartCmd(ctx, t, Target(), args...)
 	if err != nil {
		t.Fatalf("%v failed: %v", args, err)
	}

	defer func() {
		if err := ss.Stop(t); err != nil {
			t.Logf("Failed to kill mount: %v", err)
		}
	}()

	// Write local files
	want := []byte(fmt.Sprintf("test-%d\n", time.Now().UnixNano()))
	for _, name := range []string{"created-by-test", "removed-by-pod"} {
		p := filepath.Join(tempDir, name)
		err := ioutil.WriteFile(p, want, 0644)
		if err != nil {
			t.Errorf("WriteFile %s: %v", p, err)
		}
	}

	// Start the "busybox-mount" pod.
	rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "busybox-mount-test.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "integration-test=busybox-mount", 2*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}

	// Read the file written by pod startup
	p := filepath.Join(tempDir, "created-by-pod")
	got, err := ioutil.ReadFile(p)
	if err != nil {
		t.Errorf("readfile %s: %v", p, err)
	}
	want = []byte("test\n")
	if !bytes.Equal(got, want) {
		t.Errorf("%s = %q, want %q", p, got, want)
	}

	// test that file written from host was read in by the pod via cat /mount-9p/written-by-host;
	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "logs", "busybox-mount")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if !bytes.Equal(rr.Stdout.Bytes(), want) {
		t.Errorf("busybox-mount logs = %q, want %q", rr.Stdout.Bytes(), want)
	}

	// test file timestamps are correct
	for _, name := range []string{"created-by-host", "created-by-pod"} {
		gp := path.Join(guestMount, name)
		// test that file written from host was read in by the pod via cat /mount-9p/fromhost;
		rr, err := RunCmd(ctx, t, Target(), "-p", profile, "ssh", "stat", gp)
		if err != nil {
			t.Errorf("%s failed: %v", rr.Args, err)
		}
		if !bytes.Equal(rr.Stdout.Bytes(), want) {
			t.Errorf("logs = %q, want %q", rr.Stdout.Bytes(), want)
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

	p = filepath.Join(tempDir, "removed-by-pod")
	if _, err := os.Stat(p); err == nil {
		t.Errorf("expected file %s to be removed", p)
	}

	p = filepath.Join(tempDir, "created-by-pod")
	if err := os.Remove(p); err != nil {
		t.Errorf("unexpected error removing file %s: %v", p, err)
	}
}
