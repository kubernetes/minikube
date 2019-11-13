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
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	core "k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	"k8s.io/minikube/pkg/util/retry"
)

func validatePersistentVolumeClaim(ctx context.Context, t *testing.T, profile string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if _, err := PodWait(ctx, t, profile, "kube-system", "integration-test=storage-provisioner", 4*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}

	checkStorageClass := func() error {
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "storageclass", "-o=json"))
		if err != nil {
			return err
		}
		scl := storage.StorageClassList{}
		if err := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes())).Decode(&scl); err != nil {
			return err
		}
		if len(scl.Items) == 0 {
			return fmt.Errorf("no storageclass yet")
		}
		return nil
	}

	// Ensure the addon-manager has created the StorageClass before creating a claim, otherwise it won't be bound
	if err := retry.Expo(checkStorageClass, time.Second, 90*time.Second); err != nil {
		t.Errorf("no default storage class after retry: %v", err)
	}

	// Now create a testpvc
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "pvc.yaml")))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	checkStoragePhase := func() error {
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "pvc", "testpvc", "-o=json"))
		if err != nil {
			return err
		}
		pvc := core.PersistentVolumeClaim{}
		if err := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes())).Decode(&pvc); err != nil {
			return err
		}
		// The test passes if the volume claim gets bound.
		if pvc.Status.Phase == "Bound" {
			return nil
		}
		return fmt.Errorf("testpvc phase = %q, want %q (msg=%+v)", pvc.Status.Phase, "Bound", pvc)
	}

	if err := retry.Expo(checkStoragePhase, 2*time.Second, 4*time.Minute); err != nil {
		t.Fatalf("PV Creation failed with error: %v", err)
	}
}
