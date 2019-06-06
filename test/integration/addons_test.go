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
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/apimachinery/pkg/labels"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func testAddons(t *testing.T) {
	t.Parallel()
	client, err := pkgutil.GetClient()
	if err != nil {
		t.Fatalf("Could not get kubernetes client: %v", err)
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"component": "kube-addon-manager"}))
	if err := pkgutil.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
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
	minikubeRunner := NewMinikubeRunner(t)
	cmd, out := minikubeRunner.RunDaemon("dashboard --url")
	defer func() {
		err := cmd.Process.Kill()
		if err != nil {
			t.Logf("Failed to kill dashboard command: %v", err)
		}
	}()

	s, err := readLineWithTimeout(out, 180*time.Second)
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

	resp, err := http.Get(u.String())
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
	minikubeRunner := NewMinikubeRunner(t)
	kubectlRunner := util.NewKubectlRunner(t)

	minikubeRunner.RunCommand("addons enable ingress", true)
	if err := util.WaitForIngressControllerRunning(t); err != nil {
		t.Fatalf("waiting for ingress-controller to be up: %v", err)
	}

	if err := util.WaitForIngressDefaultBackendRunning(t); err != nil {
		t.Fatalf("waiting for default-http-backend to be up: %v", err)
	}

	curdir, err := filepath.Abs("")
	if err != nil {
		t.Errorf("Error getting the file path for current directory: %s", curdir)
	}
	ingressPath := path.Join(curdir, "testdata", "nginx-ing.yaml")
	if _, err := kubectlRunner.RunCommand([]string{"create", "-f", ingressPath}); err != nil {
		t.Fatalf("creating nginx ingress resource: %v", err)
	}

	podPath := path.Join(curdir, "testdata", "nginx-pod-svc.yaml")
	if _, err := kubectlRunner.RunCommand([]string{"create", "-f", podPath}); err != nil {
		t.Fatalf("creating nginx ingress resource: %v", err)
	}

	if err := util.WaitForNginxRunning(t); err != nil {
		t.Fatalf("waiting for nginx to be up: %v", err)
	}

	checkIngress := func() error {
		expectedStr := "Welcome to nginx!"
		runCmd := fmt.Sprintf("curl http://127.0.0.1:80 -H 'Host: nginx.example.com'")
		sshCmdOutput, _ := minikubeRunner.SSH(runCmd)
		if !strings.Contains(sshCmdOutput, expectedStr) {
			return fmt.Errorf("ExpectedStr sshCmdOutput to be: %s. Output was: %s", expectedStr, sshCmdOutput)
		}
		return nil
	}

	if err := util.Retry(t, checkIngress, 3*time.Second, 5); err != nil {
		t.Fatalf(err.Error())
	}

	defer func() {
		for _, p := range []string{podPath, ingressPath} {
			if out, err := kubectlRunner.RunCommand([]string{"delete", "-f", p}); err != nil {
				t.Logf("delete -f %s failed: %v\noutput: %s\n", p, err, out)
			}
		}
	}()
	minikubeRunner.RunCommand("addons disable ingress", true)
}

func testServicesList(t *testing.T) {
	t.Parallel()
	minikubeRunner := NewMinikubeRunner(t)

	checkServices := func() error {
		output := minikubeRunner.RunCommand("service list", false)
		if !strings.Contains(output, "kubernetes") {
			return fmt.Errorf("Error, kubernetes service missing from output %s", output)
		}
		return nil
	}
	if err := util.Retry(t, checkServices, 2*time.Second, 5); err != nil {
		t.Fatalf(err.Error())
	}
}

func testGvisor(t *testing.T) {
	minikubeRunner := NewMinikubeRunner(t)
	minikubeRunner.RunCommand("addons enable gvisor", true)

	t.Log("waiting for gvisor controller to come up")
	if err := util.WaitForGvisorControllerRunning(t); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	createUntrustedWorkload(t)

	t.Log("making sure untrusted workload is Running")
	if err := util.WaitForUntrustedNginxRunning(); err != nil {
		t.Fatalf("waiting for nginx to be up: %v", err)
	}

	t.Log("disabling gvisor addon")
	minikubeRunner.RunCommand("addons disable gvisor", true)
	t.Log("waiting for gvisor controller pod to be deleted")
	if err := util.WaitForGvisorControllerDeleted(); err != nil {
		t.Fatalf("waiting for gvisor controller to be deleted: %v", err)
	}

	createUntrustedWorkload(t)

	t.Log("waiting for FailedCreatePodSandBox event")
	if err := util.WaitForFailedCreatePodSandBoxEvent(); err != nil {
		t.Fatalf("waiting for FailedCreatePodSandBox event: %v", err)
	}
	deleteUntrustedWorkload(t)
}

func testGvisorRestart(t *testing.T) {
	minikubeRunner := NewMinikubeRunner(t)
	minikubeRunner.EnsureRunning()
	minikubeRunner.RunCommand("addons enable gvisor", true)

	t.Log("waiting for gvisor controller to come up")
	if err := util.WaitForGvisorControllerRunning(t); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	// TODO: @priyawadhwa to add test for stop as well
	minikubeRunner.RunCommand("delete", false)
	minikubeRunner.CheckStatus(state.None.String())
	minikubeRunner.Start()
	minikubeRunner.CheckStatus(state.Running.String())

	t.Log("waiting for gvisor controller to come up")
	if err := util.WaitForGvisorControllerRunning(t); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	createUntrustedWorkload(t)
	t.Log("making sure untrusted workload is Running")
	if err := util.WaitForUntrustedNginxRunning(); err != nil {
		t.Fatalf("waiting for nginx to be up: %v", err)
	}
	deleteUntrustedWorkload(t)
}

func createUntrustedWorkload(t *testing.T) {
	kubectlRunner := util.NewKubectlRunner(t)
	curdir, err := filepath.Abs("")
	if err != nil {
		t.Errorf("Error getting the file path for current directory: %s", curdir)
	}
	untrustedPath := path.Join(curdir, "testdata", "nginx-untrusted.yaml")
	t.Log("creating pod with untrusted workload annotation")
	if _, err := kubectlRunner.RunCommand([]string{"replace", "-f", untrustedPath, "--force"}); err != nil {
		t.Fatalf("creating untrusted nginx resource: %v", err)
	}
}

func deleteUntrustedWorkload(t *testing.T) {
	kubectlRunner := util.NewKubectlRunner(t)
	curdir, err := filepath.Abs("")
	if err != nil {
		t.Errorf("Error getting the file path for current directory: %s", curdir)
	}
	untrustedPath := path.Join(curdir, "testdata", "nginx-untrusted.yaml")
	if _, err := kubectlRunner.RunCommand([]string{"delete", "-f", untrustedPath}); err != nil {
		t.Logf("error deleting untrusted nginx resource: %v", err)
	}
}
