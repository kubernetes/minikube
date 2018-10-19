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
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func testDashboard(t *testing.T) {
	t.Parallel()
	minikubeRunner := NewMinikubeRunner(t)

	cmd, out := minikubeRunner.RunDaemon("dashboard --url")
	defer func() {
		err := cmd.Process.Kill()
		if err != nil {
			t.Logf("Failed to kill mount command: %v", err)
		}
	}()

	s, err := out.ReadString('\n')
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

	ingressPath, _ := filepath.Abs("testdata/nginx-ing.yaml")
	if _, err := kubectlRunner.RunCommand([]string{"create", "-f", ingressPath}); err != nil {
		t.Fatalf("creating nginx ingress resource: %v", err)
	}

	podPath, _ := filepath.Abs("testdata/nginx-pod-svc.yaml")
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
