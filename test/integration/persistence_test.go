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
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestPodPersistence(t *testing.T) {
	MaybeParallel(t)

	profile := Profile("persist")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	rr, err := Run(ctx, t, Target(), "start", "-p", profile)
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	rr, err = Run(ctx, t, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "busybox.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "integration-test=busybox-mount", 2*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}

	// Stop everything!
	rr, err = Run(ctx, t, Target(), "stop", "-p", profile)
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	// Start minikube, and validate that busybox is still running.
	rr, err = Run(ctx, t, Target(), "start", "-p", profile)
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "integration-test=busybox-mount", 1*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}
}
