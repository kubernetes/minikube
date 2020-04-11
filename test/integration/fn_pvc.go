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
	ctx, cancel := context.WithTimeout(ctx, Minutes(10))
	defer cancel()

	if _, err := PodWait(ctx, t, profile, "kube-system", "integration-test=storage-provisioner", Minutes(4)); err != nil {
		t.Fatalf("failed waiting for storage-provisioner: %v", err)
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
	if err := retry.Expo(checkStorageClass, time.Millisecond*500, Seconds(100)); err != nil {
		t.Errorf("failed to check for storage class: %v", err)
	}

	// Now create a testpvc
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "pvc.yaml")))
	if err != nil {
		t.Fatalf("kubectl apply pvc.yaml failed: args %q: %v", rr.Command(), err)
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

	if err := retry.Expo(checkStoragePhase, 2*time.Second, Minutes(4)); err != nil {
		t.Fatalf("failed to check storage phase: %v", err)
	}
}
