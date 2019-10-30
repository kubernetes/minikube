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
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/util/retry"
)

// TestAddons tests addons that require no special environment -- in parallel
func TestAddons(t *testing.T) {
	MaybeParallel(t)
	WaitForStartSlot(t)
	profile := UniqueProfileName("addons")
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--wait=false", "--memory=2600", "--alsologtostderr", "-v=1", "--addons=ingress", "--addons=registry"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	// Parallelized tests
	t.Run("parallel", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"Registry", validateRegistryAddon},
			{"Ingress", validateIngressAddon},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				tc.validator(ctx, t, profile)
			})
		}
	})

	// Assert that disable/enable works offline
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "addons", "enable", "dashboard", "-p", profile))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "addons", "disable", "dashboard", "-p", profile))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}

func validateIngressAddon(ctx context.Context, t *testing.T, profile string) {
	if NoneDriver() {
		t.Skipf("skipping: ssh unsupported by none")
	}

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("kubernetes client: %v", client)
	}

	if err := kapi.WaitForDeploymentToStabilize(client, "kube-system", "nginx-ingress-controller", 6*time.Minute); err != nil {
		t.Errorf("waiting for ingress-controller deployment to stabilize: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "kube-system", "app.kubernetes.io/name=nginx-ingress-controller", 8*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}

	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-ing.yaml")))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-pod-svc.yaml")))
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
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("curl http://127.0.0.1:80 -H 'Host: nginx.example.com'")))
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

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "disable", "ingress", "--alsologtostderr", "-v=1"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}

func validateRegistryAddon(ctx context.Context, t *testing.T, profile string) {
	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("kubernetes client: %v", client)
	}

	start := time.Now()
	if err := kapi.WaitForRCToStabilize(client, "kube-system", "registry", 6*time.Minute); err != nil {
		t.Errorf("waiting for registry replicacontroller to stabilize: %v", err)
	}
	t.Logf("registry stabilized in %s", time.Since(start))

	if _, err := PodWait(ctx, t, profile, "kube-system", "actual-registry=true", 6*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "kube-system", "registry-proxy=true", 6*time.Minute); err != nil {
		t.Fatalf("wait: %v", err)
	}

	// Test from inside the cluster (no curl available on busybox)
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "po", "-l", "run=registry-test", "--now"))
	if err != nil {
		t.Logf("pre-cleanup %s failed: %v (not a problem)", rr.Args, err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "run", "--rm", "registry-test", "--restart=Never", "--image=busybox", "-it", "--", "sh", "-c", "wget --spider -S http://registry.kube-system.svc.cluster.local"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	want := "HTTP/1.1 200"
	if !strings.Contains(rr.Stdout.String(), want) {
		t.Errorf("curl = %q, want *%s*", rr.Stdout.String(), want)
	}

	// Test from outside the cluster
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ip"))
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

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "disable", "registry", "--alsologtostderr", "-v=1"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}
