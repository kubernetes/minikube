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
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
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

func testAddons(t *testing.T) {
	t.Parallel()
	p := profileName(t)
	client, err := kapi.Client(p)
	if err != nil {
		t.Fatalf("Could not get kubernetes client: %v", err)
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"component": "kube-addon-manager"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		t.Errorf("Error waiting for addon manager to be up")
	}
}

func readLineWithTimeout(b *bufio.Reader, timeout time.Duration) (string, error) {
	s := make(chan string)
	e := make(chan error)
	go func() {
		read, err := b.ReadString('\n')
		if err != nil {
			e <- err
		} else {
			s <- read
		}
		close(s)
		close(e)
	}()

	select {
	case line := <-s:
		return line, nil
	case err := <-e:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout after %s", timeout)
	}
}

func testDashboard(t *testing.T) {
	t.Parallel()
	p := profileName(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")
	cmd, out := mk.RunDaemon("dashboard --url")
	defer func() {
		err := cmd.Process.Kill()
		if err != nil {
			t.Logf("Failed to kill dashboard command: %v", err)
		}
	}()

	s, err := readLineWithTimeout(out, 240*time.Second)
	if err != nil {
		t.Fatalf("failed to read url: %v", err)
	}

	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		t.Fatalf("failed to parse %q: %v", s, err)
	}

	if u.Scheme != "http" {
		t.Errorf("got Scheme %s, expected http", u.Scheme)
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("failed SplitHostPort: %v", err)
	}
	if host != "127.0.0.1" {
		t.Errorf("got host %s, expected 127.0.0.1", host)
	}

	resp, err := retryablehttp.Get(u.String())
	if err != nil {
		t.Fatalf("failed get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unable to read http response body: %v", err)
		}
		t.Errorf("%s returned status code %d, expected %d.\nbody:\n%s", u, resp.StatusCode, http.StatusOK, body)
	}
}

func testIngressController(t *testing.T) {
	t.Parallel()
	p := profileName(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")
	kr := util.NewKubectlRunner(t, p)

	mk.MustRun("addons enable ingress")
	if err := waitForIngressControllerRunning(p); err != nil {
		t.Fatalf("Failed waiting for ingress-controller to be up: %v", err)
	}

	ingressPath := filepath.Join(*testdataDir, "nginx-ing.yaml")
	if _, err := kr.RunCommand([]string{"create", "-f", ingressPath}); err != nil {
		t.Fatalf("Failed creating nginx ingress resource: %v", err)
	}

	podPath := filepath.Join(*testdataDir, "nginx-pod-svc.yaml")
	if _, err := kr.RunCommand([]string{"create", "-f", podPath}); err != nil {
		t.Fatalf("Failed creating nginx ingress resource: %v", err)
	}

	if err := waitForNginxRunning(t, p); err != nil {
		t.Fatalf("Failed waiting for nginx to be up: %v", err)
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
				t.Logf("delete -f %s failed: %v\noutput: %s\n", p, err, out)
			}
		}
	}()
	mk.MustRun("addons disable ingress")
}

func testServicesList(t *testing.T) {
	t.Parallel()
	p := profileName(t)
	mk := NewMinikubeRunner(t, p)

	checkServices := func() error {
		output, stderr, err := mk.RunCommand("service list", false)
		if err != nil {
			return err
		}
		if !strings.Contains(output, "kubernetes") {
			return fmt.Errorf("error, kubernetes service missing from output: %s, \n stderr: %s", output, stderr)
		}
		return nil
	}
	if err := retry.Expo(checkServices, 500*time.Millisecond, time.Minute); err != nil {
		t.Fatalf(err.Error())
	}
}
func testRegistry(t *testing.T) {
	t.Parallel()
	p := profileName(t)
	mk := NewMinikubeRunner(t, p)
	mk.MustRun("addons enable registry")
	client, err := kapi.Client(p)
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
	kr := util.NewKubectlRunner(t, p)
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

// waitForNginxRunning waits for nginx service to be up
func waitForNginxRunning(t *testing.T, miniProfile string) error {
	client, err := kapi.Client(miniProfile)

	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
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

// waitForIngressControllerRunning waits until ingress controller pod to be running
func waitForIngressControllerRunning(miniProfile string) error {
	client, err := kapi.Client(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	if err := kapi.WaitForDeploymentToStabilize(client, "kube-system", "nginx-ingress-controller", time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for ingress-controller deployment to stabilize")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"app.kubernetes.io/name": "nginx-ingress-controller"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "waiting for ingress-controller pods")
	}

	return nil
}
