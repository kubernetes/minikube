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

package lock

import (
	"path/filepath"
	"testing"
	"time"
)

func TestUserMutexSpec(t *testing.T) {
	var tests = []struct {
		description string
		path        string
		expected    string
	}{
		{
			description: "standard",
			path:        "/foo/bar",
		},
		{
			description: "deep directory",
			path:        "/foo/bar/baz/bat",
		},
		{
			description: "underscores",
			path:        "/foo_bar/baz",
		},
		{
			description: "starts with number",
			path:        "/foo/2bar/baz",
		},
		{
			description: "starts with punctuation",
			path:        "/.foo/bar",
		},
		{
			description: "long filename",
			path:        "/very-very-very-very-very-very-very-very-long/bar",
		},
		{
			description: "Windows kubeconfig",
			path:        `C:\Users\admin/.kube/config`,
		},
		{
			description: "Windows json",
			path:        `C:\Users\admin\.minikube\profiles\containerd-20191210T212325.7356633-8584\config.json`,
		},
	}

	seen := map[string]string{}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			got := PathMutexSpec(tc.path)
			if len(got.Name) != 40 {
				t.Errorf("%s is not 40 chars long", got.Name)
			}
			if seen[got.Name] != "" {
				t.Fatalf("lock name collision between %s and %s", tc.path, seen[got.Name])
			}
			seen[got.Name] = tc.path
			// Since we are in the lock package, we call Acquire directly
			m, err := Acquire(got)
			if err != nil {
				t.Errorf("acquire for spec %+v failed: %v", got, err)
			}
			if m != nil {
				m.Release()
			}
		})
	}
}

// TestMutexExclusion verifies that two processes cannot hold the same lock simultaneously.
// This is critical for data integrity when multiple minikube instances access shared resources.
func TestMutexExclusion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "lock")
	spec := PathMutexSpec(path)

	// 1. Acquire the lock
	r1, err := Acquire(spec)
	if err != nil {
		t.Fatalf("Failed to acquire lock 1: %v", err)
	}
	defer r1.Release()

	// 2. Try to acquire the same lock with a separate spec (simulating another process/routine)
	// We use a short timeout to fail quickly if it correctly blocks
	spec2 := PathMutexSpec(path)
	// Set a very short timeout
	spec2.Timeout = 100 * time.Millisecond
	spec2.Delay = 10 * time.Millisecond

	// This should fail because r1 is holding it
	_, err = Acquire(spec2)
	if err == nil {
		t.Fatal("Expected error acquiring locked mutex, got nil")
	}
}

// TestMutexRelease confirms that a lock can be successfully re-acquired after being released.
// This ensures that our unlocking mechanism works correctly and doesn't leave resources permanently locked.
func TestMutexRelease(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "lock")
	spec := PathMutexSpec(path)

	// 1. Acquire and Release
	r1, err := Acquire(spec)
	if err != nil {
		t.Fatalf("Failed to acquire lock 1: %v", err)
	}
	r1.Release()

	// 2. Acquire again -> should succeed
	r2, err := Acquire(spec)
	if err != nil {
		t.Fatalf("Failed to acquire lock 2 after release: %v", err)
	}
	r2.Release()
}

// TestMutexConcurrency checks that a second process will wait for an existing lock to be released rather than failing immediately.
// This ensures that multiple minikube commands can queue up safely without erroring out.
func TestMutexConcurrency(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "lock")
	spec := PathMutexSpec(path)

	// 1. Hold lock for 200ms
	started := make(chan error)
	go func() {
		r, err := Acquire(spec)
		if err != nil {
			started <- err
			return
		}
		close(started)
		time.Sleep(200 * time.Millisecond)
		r.Release()
	}()

	// Wait for the routine to acquire the lock
	if err := <-started; err != nil {
		t.Fatalf("routine 1 failed to acquire: %v", err)
	}

	// 2. Try to acquire with 500ms timeout -> should succeed eventually
	spec2 := PathMutexSpec(path)
	spec2.Timeout = 1 * time.Second
	spec2.Delay = 50 * time.Millisecond

	r2, err := Acquire(spec2)
	if err != nil {
		t.Fatalf("routine 2 failed to wait for lock: %v", err)
	}
	r2.Release()
}
