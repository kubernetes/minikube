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
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	commonutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

var (
	pvcName = "testpvc"
	pvcCmd  = []string{"get", "pvc", pvcName}
)

func testProvisioning(t *testing.T) {
	t.Parallel()
	kubectlRunner := util.NewKubectlRunner(t)

	defer func() {
		kubectlRunner.RunCommand([]string{"delete", "pvc", pvcName})
	}()

	pvcPath, _ := filepath.Abs("testdata/pvc.yaml")

	if _, err := kubectlRunner.RunCommand([]string{"create", "-f", pvcPath}); err != nil {
		t.Fatalf("Error creating pvc")
	}

	checkStorage := func() error {
		pvc := api.PersistentVolumeClaim{}
		if err := kubectlRunner.RunCommandParseOutput(pvcCmd, &pvc); err != nil {
			return &commonutil.RetriableError{Err: err}
		}
		// The test passes if the volume claim gets bound.
		if pvc.Status.Phase == "Bound" {
			return nil
		}
		return &commonutil.RetriableError{Err: fmt.Errorf("PV not attached to PVC: %v", pvc)}
	}

	if err := commonutil.RetryAfter(25, checkStorage, 20*time.Second); err != nil {
		t.Fatal("PV Creation failed with error:", err)
	}

}
