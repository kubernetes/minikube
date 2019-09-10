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
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/util/retry"
)

func TestIngressAddon(t *testing.T) {
	if NoneDriver() {
		t.Skipf("skipping: ssh unsupported by none")
	}

	profile := Profile("ingress")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer func() {
		CleanupWithLogs(t, profile, cancel)
		rr, err := Run(context.Background(), t, Target(), "-p", profile, "addons", "disable", "ingress")
		if err != nil {
			t.Logf("cleanup failed: %s: %v (probably ok)", rr.Args, err)
		}
	}()

	args := append([]string{"start", "-p", profile, "--wait=false"}, StartArgs()...)
	rr, err := Run(ctx, t, Target(), args...)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	rr, err = Run(ctx, t, Target(), "-p", profile, "addons", "enable", "ingress")
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("kubernetes client: %v", client)
	}

	if err := kapi.WaitForDeploymentToStabilize(client, "kube-system", "nginx-ingress-controller", time.Minute*5); err != nil {
		t.Errorf("waiting for ingress-controller deployment to stabilize: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "kube-system", "app.kubernetes.io/name=nginx-ingress-controller", 2*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}

	rr, err = Run(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-ing.yaml"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	rr, err = Run(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-pod-svc.yaml"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "run=nginx", 2*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if err := kapi.WaitForService(client, "default", "nginx", true, time.Millisecond*500, time.Minute*10); err != nil {
		t.Errorf("Error waiting for nginx service to be up")
	}

	want := "Welcome to nginx!"
	checkIngress := func() error {
		rr, err := Run(ctx, t, Target(), "-p", profile, "ssh", fmt.Sprintf("curl http://127.0.0.1:80 -H 'Host: nginx.example.com'"))
		if err != nil {
			return err
		}
		if rr.Stderr.String() != "" {
			t.Logf("%v: unexpected stderr: %s", rr.Args, rr.Stderr)
		}
		if !strings.Contains(rr.Stdout.String(), want) {
			return fmt.Errorf("%v stdout = %q, want %q", rr.Args, rr.Stdout, want)
		}
		return nil
	}

	if err := retry.Expo(checkIngress, 500*time.Millisecond, time.Minute); err != nil {
		t.Errorf("ingress never responded as expected on 127.0.0.1:80: %v", err)
	}

	rr, err = Run(ctx, t, Target(), "disable", "ingress")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}
