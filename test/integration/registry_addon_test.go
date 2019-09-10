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

package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/util/retry"
)

func TestRegistryAddon(t *testing.T) {
	MaybeParallel(t)
	profile := Profile("registry")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer func() {
		CleanupWithLogs(t, profile, cancel)
	}()

	args := append([]string{"start", "-p", profile, "--container-runtime=docker", "--wait=false"}, StartArgs()...)
	rr, err := Run(ctx, t, Target(), args...)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	rr, err = Run(ctx, t, Target(), "-p", profile, "addons", "enable", "registry")
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("kubernetes client: %v", client)
	}

	start := time.Now()
	if err := kapi.WaitForRCToStabilize(client, "kube-system", "registry", 3*time.Minute); err != nil {
		t.Errorf("waiting for registry replicacontroller to stabilize: %v", err)
	}
	t.Logf("registry stabilized in %s", time.Since(start))

	if _, err := PodWait(ctx, t, profile, "kube-system", "actual-registry=true", 3*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "kube-system", "registry-proxy=true", 3*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}

	// Test from inside the cluster (no curl available on busybox)
	rr, err = Run(ctx, t, "kubectl", "--context", profile, "delete", "po", "-l", "run=registry-test", "--now")
	if err != nil {
		t.Logf("pre-cleanup %s failed: %v (not a problem)", rr.Args, err)
	}

	rr, err = Run(ctx, t, "kubectl", "--context", profile, "run", "--rm", "registry-test", "--restart=Never", "--image=busybox", "-it", "--", "sh", "-c", "wget --spider -S http://registry.kube-system.svc.cluster.local")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	want := "HTTP/1.1 200"
	if !strings.Contains(rr.Stdout.String(), want) {
		t.Errorf("curl = %q, want *%s*", rr.Stdout.String(), want)
	}

	// Test from outside the cluster
	rr, err = Run(ctx, t, Target(), "-p", profile, "ip")
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	if rr.Stderr.String() != "" {
		t.Errorf("%s: unexpected stderr: %s", rr.Args, rr.Stderr)
	}

	endpoint := fmt.Sprintf("http://%s:%d", strings.TrimSpace(rr.Stdout.String()), 5000)
	u, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("failed to parse %q: %v", endpoint, err)
	}

	checkExternalAccess := func() error {
		resp, err := retryablehttp.Get(u.String())
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s = status code %d, want %d", u, resp.StatusCode, http.StatusOK)
		}
		return nil
	}

	if err := retry.Expo(checkExternalAccess, 500*time.Millisecond, 2*time.Minute); err != nil {
		t.Errorf(err.Error())
	}

	rr, err = Run(ctx, t, Target(), "addons", "disable", "registry")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}
