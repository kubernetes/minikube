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
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/util/retry"
	"k8s.io/minikube/test/integration/util"
)


func validateIngressAddon(ctx context.Context, t *testing.T, profile string) {
	MaybeParallel(t)
	client, err := kapi.Client(profile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	mk.MustRun("addons enable ingress")

	if err := kapi.WaitForDeploymentToStabilize(client, "kube-system", "nginx-ingress-controller", time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for ingress-controller deployment to stabilize")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"app.kubernetes.io/name": "nginx-ingress-controller"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "waiting for ingress-controller pods")
	}

	ingressPath := filepath.Join(*testdataDir, "nginx-ing.yaml")
	if _, err := kr.RunCommand([]string{"create", "-f", ingressPath}); err != nil {
		t.Fatalf("Failed creating nginx ingress resource: %v", err)
	}

	podPath := filepath.Join(*testdataDir, "nginx-pod-svc.yaml")
	if _, err := kr.RunCommand([]string{"create", "-f", podPath}); err != nil {
		t.Fatalf("Failed creating nginx ingress resource: %v", err)
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		return errors.Wrap(err, "waiting for nginx pods")
	}

	if err := kapi.WaitForService(client, "default", "nginx", true, time.Millisecond*500, time.Minute*10); err != nil {
		t.Errorf("Error waiting for nginx service to be up")
	}
	return nil
}

	checkIngress := func() error {
		expectedStr := "Welcome to nginx!"
		runCmd := fmt.Sprintf("curl http://127.0.0.1:80 -H 'Host: nginx.example.com'")
		sshCmdOutput, _ := mk.SSH(runCmd)
		if !strings.Contains(sshCmdOutput, expectedStr) {
			return fmt.Errorf("ExpectedStr sshCmdOutput to be: %s. Output was: %s", expectedStr, sshCmdOutput)
		}
		return nil
	}

	if err := retry.Expo(checkIngress, 500*time.Millisecond, time.Minute); err != nil {
		t.Fatalf(err.Error())
	}

	defer func() {
		for _, p := range []string{podPath, ingressPath} {
			if out, err := kr.RunCommand([]string{"delete", "-f", p}); err != nil {
				t.Logf("delete -f %s failed: %v\noutput: %s\n", profile, err, out)
			}
		}
	}()
	mk.MustRun("addons disable ingress")
}


func validateRegistryAddon(ctx context.Context, t *testing.T, profile string) {
	MaybeParallel(t)
	mk.MustRun("addons enable registry")
	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("getting kubernetes client: %v", err)
	}
	if err := kapi.WaitForRCToStabilize(client, "kube-system", "registry", time.Minute*5); err != nil {
		t.Fatalf("waiting for registry replicacontroller to stabilize: %v", err)
	}
	rs := labels.SelectorFromSet(labels.Set(map[string]string{"actual-registry": "true"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", rs); err != nil {
		t.Fatalf("waiting for registry pods: %v", err)
	}
	ps := labels.SelectorFromSet(labels.Set(map[string]string{"registry-proxy": "true"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", ps); err != nil {
		t.Fatalf("waiting for registry-proxy pods: %v", err)
	}
	ip, stderr := mk.MustRun("ip")
	ip = strings.TrimSpace(ip)
	endpoint := fmt.Sprintf("http://%s:%d", ip, 5000)
	u, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("failed to parse %q: %v stderr : %s", endpoint, err, stderr)
	}
	t.Log("checking registry access from outside cluster")

	// Check access from outside the cluster on port 5000, validing connectivity via registry-proxy
	checkExternalAccess := func() error {
		resp, err := retryablehttp.Get(u.String())
		if err != nil {
			t.Errorf("failed get: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s returned status code %d, expected %d.\n", u, resp.StatusCode, http.StatusOK)
		}
		return nil
	}

	if err := retry.Expo(checkExternalAccess, 500*time.Millisecond, 2*time.Minute); err != nil {
		t.Fatalf(err.Error())
	}
	t.Log("checking registry access from inside cluster")
	kr := util.NewKubectlRunner(t, profile)
	// TODO: Fix this
	out, _ := kr.RunCommand([]string{
		"run",
		"registry-test",
		"--restart=Never",
		"--image=busybox",
		"-it",
		"--",
		"sh",
		"-c",
		"wget --spider -S 'http://registry.kube-system.svc.cluster.local' 2>&1 | grep 'HTTP/' | awk '{print $2}'"})
	internalCheckOutput := string(out)
	expectedStr := "200"
	if !strings.Contains(internalCheckOutput, expectedStr) {
		t.Errorf("ExpectedStr internalCheckOutput to be: %s. Output was: %s", expectedStr, internalCheckOutput)
	}

	defer func() {
		if _, err := kr.RunCommand([]string{"delete", "pod", "registry-test"}); err != nil {
			t.Errorf("failed to delete pod registry-test")
		}
	}()
	mk.MustRun("addons disable registry")
}
