//go:build integration

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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	core "k8s.io/api/core/v1"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/util/retry"
)

// TestAddons tests addons that require no special environment in parallel
func TestAddons(t *testing.T) {
	profile := UniqueProfileName("addons")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer Cleanup(t, profile, cancel)

	t.Run("PreSetup", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"EnablingAddonOnNonExistingCluster", validateEnablingAddonOnNonExistingCluster},
			{"DisablingAddonOnNonExistingCluster", validateDisablingAddonOnNonExistingCluster},
		}
		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				tc.validator(ctx, t, profile)
			})
		}
	})

	setupSucceeded := t.Run("Setup", func(t *testing.T) {
		// Set an env var to point to our dummy credentials file
		// don't use t.Setenv because we sometimes manually unset the env var later manually
		err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(*testdataDir, "gcp-creds.json"))
		t.Cleanup(func() {
			os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		})
		if err != nil {
			t.Fatalf("Failed setting GOOGLE_APPLICATION_CREDENTIALS env var: %v", err)
		}

		err = os.Setenv("GOOGLE_CLOUD_PROJECT", "this_is_fake")
		t.Cleanup(func() {
			os.Unsetenv("GOOGLE_CLOUD_PROJECT")
		})
		if err != nil {
			t.Fatalf("Failed setting GOOGLE_CLOUD_PROJECT env var: %v", err)
		}

		// MOCK_GOOGLE_TOKEN forces the gcp-auth webhook to use a mock token instead of trying to get a valid one from the credentials.
		os.Setenv("MOCK_GOOGLE_TOKEN", "true")

		// for some reason, (Docker_Cloud_Shell) sets 'MINIKUBE_FORCE_SYSTEMD=true' while having cgroupfs set in docker (and probably os itself), which might make it unstable and occasionally fail:
		// - I1226 15:05:24.834294   11286 out.go:177]   - MINIKUBE_FORCE_SYSTEMD=true
		// - I1226 15:05:25.070037   11286 info.go:266] docker info: {... CgroupDriver:cgroupfs ...}
		// ref: https://storage.googleapis.com/minikube-builds/logs/15463/27154/Docker_Cloud_Shell.html
		// so we override that here to let minikube auto-detect appropriate cgroup driver
		os.Setenv(constants.MinikubeForceSystemdEnv, "")

		args := append([]string{"start", "-p", profile, "--wait=true", "--memory=4000", "--alsologtostderr", "--addons=registry", "--addons=metrics-server", "--addons=volumesnapshots", "--addons=csi-hostpath-driver", "--addons=gcp-auth", "--addons=cloud-spanner", "--addons=inspektor-gadget", "--addons=nvidia-device-plugin", "--addons=yakd", "--addons=volcano", "--addons=amd-gpu-device-plugin"}, StartArgs()...)
		if !NoneDriver() {
			args = append(args, "--addons=ingress", "--addons=ingress-dns", "--addons=storage-provisioner-rancher")
		}
		rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}

	})

	if !setupSucceeded {
		t.Fatalf("Failed setup for addon tests")
	}

	type TestCase = struct {
		name      string
		validator validateFunc
	}
	// Run tests in serial to avoid collision
	t.Run("serial", func(t *testing.T) {
		tests := []TestCase{
			{"Volcano", validateVolcanoAddon},
			{"GCPAuth", validateGCPAuthAddon},
		}
		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
			})
		}
	})

	// Parallelized tests
	t.Run("parallel", func(t *testing.T) {
		tests := []TestCase{
			{"Registry", validateRegistryAddon},
			{"Ingress", validateIngressAddon},
			{"InspektorGadget", validateInspektorGadgetAddon},
			{"MetricsServer", validateMetricsServerAddon},
			{"Olm", validateOlmAddon},
			{"CSI", validateCSIDriverAndSnapshots},
			{"Headlamp", validateHeadlampAddon},
			{"CloudSpanner", validateCloudSpannerAddon},
			{"LocalPath", validateLocalPathAddon},
			{"NvidiaDevicePlugin", validateNvidiaDevicePlugin},
			{"Yakd", validateYakdAddon},
			{"AmdGpuDevicePlugin", validateAmdGpuDevicePlugin},
		}

		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				tc.validator(ctx, t, profile)
			})
		}
	})

	t.Run("StoppedEnableDisable", func(t *testing.T) {
		// Assert that disable/enable works offline
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile))
		if err != nil {
			t.Errorf("failed to stop minikube. args %q : %v", rr.Command(), err)
		}
		rr, err = Run(t, exec.CommandContext(ctx, Target(), "addons", "enable", "dashboard", "-p", profile))
		if err != nil {
			t.Errorf("failed to enable dashboard addon: args %q : %v", rr.Command(), err)
		}
		rr, err = Run(t, exec.CommandContext(ctx, Target(), "addons", "disable", "dashboard", "-p", profile))
		if err != nil {
			t.Errorf("failed to disable dashboard addon: args %q : %v", rr.Command(), err)
		}
		// Disable a non-enabled addon
		rr, err = Run(t, exec.CommandContext(ctx, Target(), "addons", "disable", "gvisor", "-p", profile))
		if err != nil {
			t.Errorf("failed to disable non-enabled addon: args %q : %v", rr.Command(), err)
		}
	})
}

// validateIngressAddon tests the ingress addon by deploying a default nginx pod
func validateIngressAddon(ctx context.Context, t *testing.T, profile string) {
	if NoneDriver() {
		t.Skipf("skipping: ingress not supported")
	}
	defer disableAddon(t, "ingress", profile)
	defer disableAddon(t, "ingress-dns", profile)
	defer PostMortemLogs(t, profile)

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("failed to get Kubernetes client: %v", client)
	}

	// avoid timeouts like:
	// Error from server (InternalError): Internal error occurred: failed calling webhook "validate.nginx.ingress.kubernetes.io": Post "https://ingress-nginx-controller-admission.ingress-nginx.svc:443/networking/v1/ingresses?timeout=10s": dial tcp 10.107.218.58:443: i/o timeout
	// Error from server (InternalError): Internal error occurred: failed calling webhook "validate.nginx.ingress.kubernetes.io": Post "https://ingress-nginx-controller-admission.ingress-nginx.svc:443/networking/v1/ingresses?timeout=10s": context deadline exceeded
	if _, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "wait", "--for=condition=ready", "--namespace=ingress-nginx", "pod", "--selector=app.kubernetes.io/component=controller", "--timeout=90s")); err != nil {
		t.Fatalf("failed waiting for ingress-nginx-controller : %v", err)
	}

	// use nginx ingress yaml that corresponds to k8s version
	// default: k8s >= v1.19, ingress api v1
	ingressYaml := "nginx-ingress-v1.yaml"
	ingressDNSYaml := "ingress-dns-example-v1.yaml"
	v, err := client.ServerVersion()
	if err == nil {
		// for pre-release k8s version, remove any "+" suffix in minor version to be semver-compliant and not panic
		// ref: https://github.com/kubernetes/minikube/pull/16145#issuecomment-1483283260
		minor := strings.TrimSuffix(v.Minor, "+")
		if semver.MustParseRange("<1.19.0")(semver.MustParse(fmt.Sprintf("%s.%s.0", v.Major, minor))) {
			// legacy: k8s < v1.19 & ingress api v1beta1
			ingressYaml = "nginx-ingress-v1beta1.yaml"
			ingressDNSYaml = "ingress-dns-example-v1beta1.yaml"
		}
	} else {
		t.Log("failed to get k8s version, assuming v1.19+ => ingress api v1")
	}

	// create networking.k8s.io/v1 ingress
	createv1Ingress := func() error {
		// apply networking.k8s.io/v1 ingress
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, ingressYaml)))
		if err != nil {
			return err
		}
		if rr.Stderr.String() != "" {
			t.Logf("%v: unexpected stderr: %s (may be temporary)", rr.Command(), rr.Stderr)
		}
		return nil
	}
	if err := retry.Expo(createv1Ingress, 1*time.Second, Seconds(90)); err != nil {
		t.Errorf("failed to create ingress: %v", err)
	}

	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-pod-svc.yaml")))
	if err != nil {
		t.Errorf("failed to kubectl replace nginx-pod-svc. args %q. %v", rr.Command(), err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "run=nginx", Minutes(8)); err != nil {
		t.Fatalf("failed waiting for ngnix pod: %v", err)
	}
	if err := kapi.WaitForService(client, "default", "nginx", true, time.Millisecond*500, Minutes(10)); err != nil {
		t.Errorf("failed waiting for nginx service to be up: %v", err)
	}

	want := "Welcome to nginx!"
	addr := "http://127.0.0.1/"

	// check if the ingress can route nginx app with networking.k8s.io/v1 ingress
	checkv1Ingress := func() error {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("curl -s %s -H 'Host: nginx.example.com'", addr)))
		if err != nil {
			return err
		}

		stderr := rr.Stderr.String()
		if rr.Stderr.String() != "" {
			t.Logf("debug: unexpected stderr for %v:\n%s", rr.Command(), stderr)
		}
		stdout := rr.Stdout.String()
		if !strings.Contains(stdout, want) {
			return fmt.Errorf("%v stdout = %q, want %q", rr.Command(), stdout, want)
		}
		return nil
	}
	if err := retry.Expo(checkv1Ingress, 500*time.Millisecond, Seconds(90)); err != nil {
		t.Errorf("failed to get expected response from %s within minikube: %v", addr, err)
	}

	if NeedsPortForward() {
		t.Skip("skipping ingress DNS test for any combination that needs port forwarding")
	}

	// check the ingress-dns addon here as well
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, ingressDNSYaml)))
	if err != nil {
		t.Errorf("failed to kubectl replace ingress-dns-example. args %q. %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ip"))
	if err != nil {
		t.Errorf("failed to retrieve minikube ip. args %q : %v", rr.Command(), err)
	}
	ip := strings.TrimSuffix(rr.Stdout.String(), "\n")

	rr, err = Run(t, exec.CommandContext(ctx, "nslookup", "hello-john.test", ip))
	if err != nil {
		t.Errorf("failed to nslookup hello-john.test host. args %q : %v", rr.Command(), err)
	}
	// nslookup should include info about the hello-john.test host, including minikube's ip
	if !strings.Contains(rr.Stdout.String(), ip) {
		t.Errorf("unexpected output from nslookup. stdout: %v\nstderr: %v", rr.Stdout.String(), rr.Stderr.String())
	}
}

// validateRegistryAddon tests the registry addon
func validateRegistryAddon(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "registry", profile)
	defer PostMortemLogs(t, profile)

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("failed to get Kubernetes client for %s : %v", profile, err)
	}

	start := time.Now()
	if err := kapi.WaitForDeploymentToStabilize(client, "kube-system", "registry", Minutes(6)); err != nil {
		t.Errorf("failed waiting for registry deployment to stabilize: %v", err)
	}
	t.Logf("registry stabilized in %s", time.Since(start))

	if _, err := PodWait(ctx, t, profile, "kube-system", "actual-registry=true", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for pod actual-registry: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "kube-system", "registry-proxy=true", Minutes(10)); err != nil {
		t.Fatalf("failed waiting for pod registry-proxy: %v", err)
	}

	// Test from inside the cluster (no curl available on busybox)
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "po", "-l", "run=registry-test", "--now"))
	if err != nil {
		t.Logf("pre-cleanup %s failed: %v (not a problem)", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "run", "--rm", "registry-test", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "sh", "-c", "wget --spider -S http://registry.kube-system.svc.cluster.local"))
	if err != nil {
		t.Errorf("failed to hit registry.kube-system.svc.cluster.local. args %q failed: %v", rr.Command(), err)
	}
	want := "HTTP/1.1 200"
	if !strings.Contains(rr.Stdout.String(), want) {
		t.Errorf("expected curl response be %q, but got *%s*", want, rr.Stdout.String())
	}

	if NeedsPortForward() {
		t.Skipf("Unable to complete rest of the test due to connectivity assumptions")
	}

	// Test from outside the cluster
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ip"))
	if err != nil {
		t.Fatalf("failed run minikube ip. args %q : %v", rr.Command(), err)
	}
	if rr.Stderr.String() != "" {
		t.Errorf("expected stderr to be -empty- but got: *%q* .  args %q", rr.Stderr, rr.Command())
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

	if err := retry.Expo(checkExternalAccess, 500*time.Millisecond, Seconds(150)); err != nil {
		t.Errorf("failed to check external access to %s: %v", u.String(), err.Error())
	}
}

// validateMetricsServerAddon tests the metrics server addon by making sure "kubectl top pods" returns a sensible result
func validateMetricsServerAddon(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "metrics-server", profile)
	defer PostMortemLogs(t, profile)

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("failed to get Kubernetes client for %s: %v", profile, err)
	}

	start := time.Now()
	if err := kapi.WaitForDeploymentToStabilize(client, "kube-system", "metrics-server", Minutes(6)); err != nil {
		t.Errorf("failed waiting for metrics-server deployment to stabilize: %v", err)
	}
	t.Logf("metrics-server stabilized in %s", time.Since(start))

	if _, err := PodWait(ctx, t, profile, "kube-system", "k8s-app=metrics-server", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for k8s-app=metrics-server pod: %v", err)
	}

	want := "CPU(cores)"
	checkMetricsServer := func() error {
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "top", "pods", "-n", "kube-system"))
		if err != nil {
			return err
		}
		if rr.Stderr.String() != "" {
			t.Logf("%v: unexpected stderr: %s", rr.Command(), rr.Stderr)
		}
		if !strings.Contains(rr.Stdout.String(), want) {
			return fmt.Errorf("%v stdout = %q, want %q", rr.Command(), rr.Stdout, want)
		}
		return nil
	}

	if err := retry.Expo(checkMetricsServer, time.Second*3, Minutes(6)); err != nil {
		t.Errorf("failed checking metric server: %v", err.Error())
	}
}

// validateOlmAddon tests the OLM addon
func validateOlmAddon(ctx context.Context, t *testing.T, profile string) {
	t.Skip("Skipping OLM addon test until https://github.com/operator-framework/operator-lifecycle-manager/issues/2534 is resolved")
	defer disableAddon(t, "olm", profile)
	defer PostMortemLogs(t, profile)
	start := time.Now()

	if _, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "wait", "--for=condition=ready", "--namespace=olm", "pod", "--selector=app=catalog-operator", "--timeout=90s")); err != nil {
		t.Fatalf("failed waiting for pod catalog-operator: %v", err)
	}
	t.Logf("catalog-operator stabilized in %s", time.Since(start))

	if _, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "wait", "--for=condition=ready", "--namespace=olm", "pod", "--selector=app=olm-operator", "--timeout=90s")); err != nil {
		t.Fatalf("failed waiting for pod olm-operator: %v", err)
	}
	t.Logf("olm-operator stabilized in %s", time.Since(start))

	if _, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "wait", "--for=condition=ready", "--namespace=olm", "pod", "--selector=app=packageserver", "--timeout=90s")); err != nil {
		t.Fatalf("failed waiting for pod olm-operator: %v", err)
	}
	t.Logf("packageserver stabilized in %s", time.Since(start))

	if _, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "wait", "--for=condition=ready", "--namespace=olm", "pod", "--selector=olm.catalogSource=operatorhubio-catalog", "--timeout=90s")); err != nil {
		t.Fatalf("failed waiting for pod operatorhubio-catalog: %v", err)
	}
	t.Logf("operatorhubio-catalog stabilized in %s", time.Since(start))

	// Install one sample Operator such as etcd
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "etcd.yaml")))
	if err != nil {
		t.Logf("etcd operator installation with %s failed: %v", rr.Command(), err)
	}

	want := "Succeeded"
	checkOperatorInstalled := func() error {
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "csv", "-n", "my-etcd"))
		if err != nil {
			return err
		}
		if rr.Stderr.String() != "" {
			t.Logf("%v: unexpected stderr: %s", rr.Command(), rr.Stderr)
		}
		if !strings.Contains(rr.Stdout.String(), want) {
			return fmt.Errorf("%v stdout = %q, want %q", rr.Command(), rr.Stdout, want)
		}
		return nil
	}
	// Operator installation takes a while
	if err := retry.Expo(checkOperatorInstalled, time.Second*3, Minutes(10)); err != nil {
		t.Errorf("failed checking operator installed: %v", err.Error())
	}
}

// validateCSIDriverAndSnapshots tests the csi hostpath driver by creating a persistent volume, snapshotting it and restoring it.
func validateCSIDriverAndSnapshots(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "csi-hostpath-driver", profile)
	defer disableAddon(t, "volumesnapshots", profile)
	defer PostMortemLogs(t, profile)

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("failed to get Kubernetes client for %s: %v", profile, err)
	}

	start := time.Now()
	if err := kapi.WaitForPods(client, "kube-system", "kubernetes.io/minikube-addons=csi-hostpath-driver", Minutes(6)); err != nil {
		t.Errorf("failed waiting for csi-hostpath-driver pods to stabilize: %v", err)
	}
	t.Logf("csi-hostpath-driver pods stabilized in %s", time.Since(start))

	// create sample PVC
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "csi-hostpath-driver", "pvc.yaml")))
	if err != nil {
		t.Logf("creating sample PVC with %s failed: %v", rr.Command(), err)
	}

	if err := PVCWait(ctx, t, profile, "default", "hpvc", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for PVC hpvc: %v", err)
	}

	// create sample pod with the PVC
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "csi-hostpath-driver", "pv-pod.yaml")))
	if err != nil {
		t.Logf("creating pod with %s failed: %v", rr.Command(), err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "app=task-pv-pod", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for pod task-pv-pod: %v", err)
	}

	// create volume snapshot
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "csi-hostpath-driver", "snapshot.yaml")))
	if err != nil {
		t.Logf("creating pod with %s failed: %v", rr.Command(), err)
	}

	if err := VolumeSnapshotWait(ctx, t, profile, "default", "new-snapshot-demo", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for volume snapshot new-snapshot-demo: %v", err)
	}

	// delete pod
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "pod", "task-pv-pod"))
	if err != nil {
		t.Logf("deleting pod with %s failed: %v", rr.Command(), err)
	}

	// delete pvc
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "pvc", "hpvc"))
	if err != nil {
		t.Logf("deleting pod with %s failed: %v", rr.Command(), err)
	}

	// restore pv from snapshot
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "csi-hostpath-driver", "pvc-restore.yaml")))
	if err != nil {
		t.Logf("creating pvc with %s failed: %v", rr.Command(), err)
	}

	if err = PVCWait(ctx, t, profile, "default", "hpvc-restore", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for PVC hpvc-restore: %v", err)
	}

	// create pod from restored snapshot
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "csi-hostpath-driver", "pv-pod-restore.yaml")))
	if err != nil {
		t.Logf("creating pod with %s failed: %v", rr.Command(), err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "app=task-pv-pod-restore", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for pod task-pv-pod-restore: %v", err)
	}

	// CLEANUP
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "pod", "task-pv-pod-restore"))
	if err != nil {
		t.Logf("cleanup with %s failed: %v", rr.Command(), err)
	}
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "pvc", "hpvc-restore"))
	if err != nil {
		t.Logf("cleanup with %s failed: %v", rr.Command(), err)
	}
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "volumesnapshot", "new-snapshot-demo"))
	if err != nil {
		t.Logf("cleanup with %s failed: %v", rr.Command(), err)
	}
}

// validateGCPAuthNamespaces validates that newly created namespaces contain the gcp-auth secret.
func validateGCPAuthNamespaces(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "ns", "new-namespace"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	logsAsError := func() error {
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "logs", "-l", "app=gcp-auth", "-n", "gcp-auth"))
		if err != nil {
			return err
		}
		return errors.New(rr.Output())
	}

	getSecret := func() error {
		_, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "secret", "gcp-auth", "-n", "new-namespace"))
		if err != nil {
			err = fmt.Errorf("%w: gcp-auth container logs: %v", err, logsAsError())
		}
		return err
	}

	if err := retry.Expo(getSecret, Seconds(2), Minutes(1)); err != nil {
		t.Errorf("failed to get secret: %v", err)
	}
}

// validateGCPAuthAddon tests the GCP Auth addon with either phony or real credentials and makes sure the files are mounted into pods correctly
func validateGCPAuthAddon(ctx context.Context, t *testing.T, profile string) {
	defer func() {
		disableGCPAuth := func() error {
			disableAddon(t, "gcp-auth", profile)
			return nil
		}
		if err := retry.Expo(disableGCPAuth, Minutes(2), Minutes(10), 5); err != nil {
			t.Errorf("failed to disable GCP auth addon: %v", err)
		}
	}()
	defer PostMortemLogs(t, profile)

	t.Run("Namespaces", func(t *testing.T) {
		validateGCPAuthNamespaces(ctx, t, profile)
	})

	t.Run("FakeCredentials", func(t *testing.T) {
		// schedule a pod to check environment variables
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "busybox.yaml")))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}

		serviceAccountName := "gcp-auth-test"
		// create a dummy service account so we know the pull secret got added
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "sa", serviceAccountName))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}

		// 8 minutes, because 4 is not enough for images to pull in all cases.
		names, err := PodWait(ctx, t, profile, "default", "integration-test=busybox", Minutes(8))
		if err != nil {
			t.Fatalf("wait: %v", err)
		}

		// Use this pod to confirm that the env vars are set correctly
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", names[0], "--", "/bin/sh", "-c", "printenv GOOGLE_APPLICATION_CREDENTIALS"))
		if err != nil {
			t.Fatalf("printenv creds: %v", err)
		}

		got := strings.TrimSpace(rr.Stdout.String())
		expected := "/google-app-creds.json"
		if got != expected {
			t.Errorf("'printenv GOOGLE_APPLICATION_CREDENTIALS' returned %s, expected %s", got, expected)
		}

		// Now check the service account and make sure the "gcp-auth" image pull secret is present
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "describe", "sa", serviceAccountName))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}

		expectedPullSecret := "gcp-auth"
		re := regexp.MustCompile(`.*Image pull secrets:.*`)
		secrets := re.FindString(rr.Stdout.String())
		if !strings.Contains(secrets, expectedPullSecret) {
			t.Errorf("Unexpected image pull secrets. expected %s, got %s", expectedPullSecret, secrets)
		}

		if !detect.IsOnGCE() || detect.IsCloudShell() {
			// Make sure the file contents are correct
			rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", names[0], "--", "/bin/sh", "-c", "cat /google-app-creds.json"))
			if err != nil {
				t.Fatalf("cat creds: %v", err)
			}

			var gotJSON map[string]string
			err = json.Unmarshal(bytes.TrimSpace(rr.Stdout.Bytes()), &gotJSON)
			if err != nil {
				t.Fatalf("unmarshal json: %v", err)
			}
			expectedJSON := map[string]string{
				"client_id":        "haha",
				"client_secret":    "nice_try",
				"quota_project_id": "this_is_fake",
				"refresh_token":    "maybe_next_time",
				"type":             "authorized_user",
			}

			if !reflect.DeepEqual(gotJSON, expectedJSON) {
				t.Fatalf("unexpected creds file: got %v, expected %v", gotJSON, expectedJSON)
			}
		}

		// Check the GOOGLE_CLOUD_PROJECT env var as well
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", names[0], "--", "/bin/sh", "-c", "printenv GOOGLE_CLOUD_PROJECT"))
		if err != nil {
			t.Fatalf("print env project: %v", err)
		}

		got = strings.TrimSpace(rr.Stdout.String())
		expected = "this_is_fake"

		if got != expected {
			t.Errorf("'printenv GOOGLE_CLOUD_PROJECT' returned %s, expected %s", got, expected)
		}
	})

	t.Run("RealCredentials", func(t *testing.T) {
		if !detect.IsOnGCE() || detect.IsCloudShell() || VMDriver() {
			t.Skip("This test requires a GCE instance (excluding Cloud Shell) with a container based driver")
		}

		// If we're on GCE, we have proper credentials and can test the registry secrets with an artifact registry image
		t.Skip("skipping GCPAuth addon test until 'Permission \"artifactregistry.repositories.downloadArtifacts\" denied on resource \"projects/k8s-minikube/locations/us/repositories/test-artifacts\" (or it may not exist)' issue is resolved")
		// "Setting the environment variable MOCK_GOOGLE_TOKEN to true will prevent using the google application credentials to fetch the token used for the image pull secret. Instead the token will be mocked."
		// ref: https://github.com/GoogleContainerTools/gcp-auth-webhook#gcp-auth-webhook
		os.Unsetenv("MOCK_GOOGLE_TOKEN")
		// re-set MOCK_GOOGLE_TOKEN once we're done
		defer os.Setenv("MOCK_GOOGLE_TOKEN", "true")

		os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		os.Unsetenv("GOOGLE_CLOUD_PROJECT")
		args := []string{"-p", profile, "addons", "enable", "gcp-auth"}
		rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
		if err != nil {
			t.Errorf("%s failed: %v", rr.Command(), err)
		} else if !strings.Contains(rr.Output(), "It seems that you are running in GCE") {
			t.Errorf("Unexpected error message: %v", rr.Output())
		}
		_, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "private-image.yaml")))
		if err != nil {
			t.Fatalf("print env project: %v", err)
		}

		// Make sure the pod is up and running, which means we successfully pulled the private image down
		// 8 minutes, because 4 is not enough for images to pull in all cases.
		_, err = PodWait(ctx, t, profile, "default", "integration-test=private-image", Minutes(8))
		if err != nil {
			t.Fatalf("wait for private image: %v", err)
		}

		// Try it with a European mirror as well
		_, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "private-image-eu.yaml")))
		if err != nil {
			t.Fatalf("print env project: %v", err)
		}

		_, err = PodWait(ctx, t, profile, "default", "integration-test=private-image-eu", Minutes(8))
		if err != nil {
			t.Fatalf("wait for private image: %v", err)
		}
	})
}

func validateHeadlampAddon(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "headlamp", profile)
	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), "addons", "enable", "headlamp", "-p", profile, "--alsologtostderr", "-v=1"))
	if err != nil {
		t.Fatalf("failed to enable headlamp addon: args: %q: %v", rr.Command(), err)
	}

	if _, err := PodWait(ctx, t, profile, "headlamp", "app.kubernetes.io/name=headlamp", Minutes(8)); err != nil {
		t.Fatalf("failed waiting for headlamp pod: %v", err)
	}
}

// validateInspektorGadgetAddon tests the inspektor-gadget addon by ensuring the pod has come up and addon disables
func validateInspektorGadgetAddon(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "inspektor-gadget", profile)
	defer PostMortemLogs(t, profile)

	if _, err := PodWait(ctx, t, profile, "gadget", "k8s-app=gadget", Minutes(8)); err != nil {
		t.Fatalf("failed waiting for inspektor-gadget pod: %v", err)
	}
}

// validateCloudSpannerAddon tests the cloud-spanner addon by ensuring the deployment and pod come up and addon disables
func validateCloudSpannerAddon(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "cloud-spanner", profile)
	defer PostMortemLogs(t, profile)

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("failed to get Kubernetes client for %s: %v", profile, err)
	}
	if err := kapi.WaitForDeploymentToStabilize(client, "default", "cloud-spanner-emulator", Minutes(6)); err != nil {
		t.Errorf("failed waiting for cloud-spanner-emulator deployment to stabilize: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "default", "app=cloud-spanner-emulator", Minutes(6)); err != nil {
		t.Errorf("failed waiting for app=cloud-spanner-emulator pod: %v", err)
	}
}

// validateVolcanoAddon tests the Volcano addon, makes sure the Volcano is installed into cluster.
func validateVolcanoAddon(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "volcano", profile)
	defer PostMortemLogs(t, profile)
	if ContainerRuntime() == "crio" {
		t.Skipf("skipping: crio not supported")
	}

	volcanoNamespace := "volcano-system"

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("failed to get Kubernetes client for %s: %v", profile, err)
	}

	// Wait for the volcano component installation to complete
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		if err := kapi.WaitForDeploymentToStabilize(client, volcanoNamespace, "volcano-scheduler", Minutes(6)); err != nil {
			t.Errorf("failed waiting for volcano-scheduler deployment to stabilize: %v", err)
		} else {
			t.Logf("volcano-scheduler stabilized in %s", time.Since(start))
		}
		wg.Done()
	}()
	go func() {
		if err := kapi.WaitForDeploymentToStabilize(client, volcanoNamespace, "volcano-admission", Minutes(6)); err != nil {
			t.Errorf("failed waiting for volcano-admission deployment to stabilize: %v", err)
		} else {
			t.Logf("volcano-admission stabilized in %s", time.Since(start))
		}
		wg.Done()
	}()
	go func() {
		if err := kapi.WaitForDeploymentToStabilize(client, volcanoNamespace, "volcano-controllers", Minutes(6)); err != nil {
			t.Errorf("failed waiting for volcano-controller deployment to stabilize: %v", err)
		} else {
			t.Logf("volcano-controller stabilized in %s", time.Since(start))
		}
		wg.Done()
	}()
	wg.Wait()

	if _, err := PodWait(ctx, t, profile, volcanoNamespace, "app=volcano-scheduler", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for app=volcano-scheduler pod: %v", err)
	}

	if _, err := PodWait(ctx, t, profile, volcanoNamespace, "app=volcano-admission", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for app=volcano-admission pod: %v", err)
	}

	if _, err := PodWait(ctx, t, profile, volcanoNamespace, "app=volcano-controller", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for app=volcano-controller pod: %v", err)
	}

	// When the volcano deployment is complete, delete the volcano-admission-init job, it will affect the tests
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "-n", volcanoNamespace, "job", "volcano-admission-init"))
	if err != nil {
		t.Logf("vcjob creation with %s failed: %v", rr.Command(), err)
	}

	// Create a vcjob
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "vcjob.yaml")))
	if err != nil {
		t.Logf("vcjob creation with %s failed: %v", rr.Command(), err)
	}

	want := "test-job"
	checkVolcano := func() error {
		// check the vcjob
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "vcjob", "-n", "my-volcano"))
		if err != nil {
			return err
		}
		if rr.Stderr.String() != "" {
			t.Logf("%v: unexpected stderr: %s", rr.Command(), rr.Stderr)
		}
		if !strings.Contains(rr.Stdout.String(), want) {
			return fmt.Errorf("%v stdout = %q, want %q", rr.Command(), rr.Stdout, want)
		}
		return nil
	}

	if err := retry.Expo(checkVolcano, time.Second*3, Minutes(2)); err != nil {
		t.Errorf("failed checking volcano: %v", err.Error())
	}

	// also ensure the job is actually running
	if _, err := PodWait(ctx, t, profile, "my-volcano", "volcano.sh/job-name=test-job", Minutes(3)); err != nil {
		t.Fatalf("failed waiting for test-local-path pod: %v", err)
	}
}

// validateLocalPathAddon tests the functionality of the storage-provisioner-rancher addon
func validateLocalPathAddon(ctx context.Context, t *testing.T, profile string) {
	if NoneDriver() {
		t.Skipf("skip local-path test on none driver")
	}
	defer disableAddon(t, "storage-provisioner-rancher", profile)
	defer PostMortemLogs(t, profile)

	// Create a test PVC
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "storage-provisioner-rancher", "pvc.yaml")))
	if err != nil {
		t.Fatalf("kubectl apply pvc.yaml failed: args %q: %v", rr.Command(), err)
	}

	// Deploy a simple pod with PVC
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "storage-provisioner-rancher", "pod.yaml")))
	if err != nil {
		t.Fatalf("kubectl apply pod.yaml failed: args %q: %v", rr.Command(), err)
	}
	if err := PVCWait(ctx, t, profile, "default", "test-pvc", Minutes(5)); err != nil {
		t.Fatalf("failed waiting for PVC test-pvc: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "default", "run=test-local-path", Minutes(3)); err != nil {
		t.Fatalf("failed waiting for test-local-path pod: %v", err)
	}

	// Get info about PVC
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "pvc", "test-pvc", "-o=json"))
	if err != nil {
		t.Fatalf("kubectl get pvc with %s failed: %v", rr.Command(), err)
	}
	pvc := core.PersistentVolumeClaim{}
	if err := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes())).Decode(&pvc); err != nil {
		t.Fatalf("failed decoding json to pvc: %v", err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("cat /opt/local-path-provisioner/%s_default_test-pvc/file1", pvc.Spec.VolumeName)))
	if err != nil {
		t.Fatalf("ssh error: %v", err)
	}

	got := rr.Stdout.String()
	want := "local-path-provisioner"
	if !strings.Contains(got, want) {
		t.Fatalf("%v stdout = %q, want %q", rr.Command(), got, want)
	}

	// Cleanup
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "pod", "test-local-path"))
	if err != nil {
		t.Logf("cleanup with %s failed: %v", rr.Command(), err)
	}
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "pvc", "test-pvc"))
	if err != nil {
		t.Logf("cleanup with %s failed: %v", rr.Command(), err)
	}
}

// validateEnablingAddonOnNonExistingCluster tests enabling an addon on a non-existing cluster
func validateEnablingAddonOnNonExistingCluster(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "addons", "enable", "dashboard", "-p", profile))
	if err == nil {
		t.Fatalf("enabling addon succeeded when it shouldn't have: %s", rr.Output())
	}
	if !strings.Contains(rr.Output(), "To start a cluster, run") {
		t.Fatalf("unexpected error was returned: %v", err)
	}
}

// validateDisablingAddonOnNonExistingCluster tests disabling an addon on a non-existing cluster
func validateDisablingAddonOnNonExistingCluster(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "addons", "disable", "dashboard", "-p", profile))
	if err == nil {
		t.Fatalf("disabling addon succeeded when it shouldn't have: %s", rr.Output())
	}
	if !strings.Contains(rr.Output(), "To start a cluster, run") {
		t.Fatalf("unexpected error was returned: %v", err)
	}
}

// validateNvidiaDevicePlugin tests the nvidia-device-plugin addon by ensuring the pod comes up and the addon disables
func validateNvidiaDevicePlugin(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "nvidia-device-plugin", profile)
	defer PostMortemLogs(t, profile)

	if _, err := PodWait(ctx, t, profile, "kube-system", "name=nvidia-device-plugin-ds", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for nvidia-device-plugin-ds pod: %v", err)
	}
}

// validateAmdGpuDevicePlugin tests the amd-gpu-device-plugin addon by ensuring the pod comes up and the addon disables
func validateAmdGpuDevicePlugin(ctx context.Context, t *testing.T, profile string) {
	if !(DockerDriver() && amd64Platform()) {
		t.Skipf("skip amd gpu test on all but docker driver and amd64 platform")
	}
	defer disableAddon(t, "amd-gpu-device-plugin", profile)
	defer PostMortemLogs(t, profile)

	if _, err := PodWait(ctx, t, profile, "kube-system", "name=amd-gpu-device-plugin", Minutes(6)); err != nil {
		t.Fatalf("failed waiting for amd-gpu-device-plugin pod: %v", err)
	}
}

func validateYakdAddon(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "yakd", profile)
	defer PostMortemLogs(t, profile)

	if _, err := PodWait(ctx, t, profile, "yakd-dashboard", "app.kubernetes.io/name=yakd-dashboard", Minutes(2)); err != nil {
		t.Fatalf("failed waiting for YAKD - Kubernetes Dashboard pod: %v", err)
	}
}

func disableAddon(t *testing.T, addon, profile string) {
	rr, err := Run(t, exec.Command(Target(), "-p", profile, "addons", "disable", addon, "--alsologtostderr", "-v=1"))
	if err != nil {
		t.Errorf("failed to disable %s addon: args %q: %v", addon, rr.Command(), err)
	}
}
