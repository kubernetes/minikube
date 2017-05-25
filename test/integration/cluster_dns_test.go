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
	"strings"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	commonutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func testClusterDNS(t *testing.T) {
	t.Parallel()
	kubectlRunner := util.NewKubectlRunner(t)
	podName := "busybox"
	podPath, _ := filepath.Abs("testdata/busybox.yaml")
	defer kubectlRunner.RunCommand([]string{"delete", "-f", podPath})

	setupTest := func() error {
		if _, err := kubectlRunner.RunCommand([]string{"create", "-f", podPath}); err != nil {
			return err
		}
		return nil
	}

	if err := commonutil.RetryAfter(20, setupTest, 2*time.Second); err != nil {
		t.Fatal("Error setting up DNS test.")
	}

	dnsTest := func() error {
		p := &api.Pod{}
		for p.Status.Phase != "Running" {
			var err error
			p, err = kubectlRunner.GetPod(podName, "default")
			if err != nil {
				return &commonutil.RetriableError{Err: err}
			}
		}

		dnsByteArr, err := kubectlRunner.RunCommand([]string{"exec", podName,
			"nslookup", "kubernetes"})
		dnsOutput := string(dnsByteArr)
		if err != nil {
			return &commonutil.RetriableError{Err: err}
		}

		if !strings.Contains(dnsOutput, "10.0.0.1") || !strings.Contains(dnsOutput, "10.0.0.10") {
			return fmt.Errorf("DNS lookup failed, could not find both 10.0.0.1 and 10.0.0.10.  Output: %s", dnsOutput)
		}
		return nil
	}

	if err := commonutil.RetryAfter(40, dnsTest, 5*time.Second); err != nil {
		t.Fatal("DNS lookup failed with error:", err)
	}
}
