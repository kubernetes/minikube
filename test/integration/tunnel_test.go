/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/minikube/pkg/minikube/tunnel"
	commonutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func testTunnel(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Otherwise minikube fails waiting for a password.
		if err := exec.Command("sudo", "-n", "route").Run(); err != nil {
			t.Skipf("password required to execute 'route', skipping testTunnel: %v", err)
		}
	}

	t.Log("starting tunnel test...")
	runner := NewMinikubeRunner(t)
	go func() {
		output := runner.RunCommand("tunnel --alsologtostderr -v 8 --logtostderr", true)
		if t.Failed() {
			fmt.Println(output)
		}
	}()

	err := tunnel.NewManager().CleanupNotRunningTunnels()

	if err != nil {
		t.Fatal(errors.Wrap(err, "cleaning up tunnels"))
	}

	kubectlRunner := util.NewKubectlRunner(t)

	t.Log("deploying nginx...")
	podPath, _ := filepath.Abs("testdata/testsvc.yaml")
	if _, err := kubectlRunner.RunCommand([]string{"apply", "-f", podPath}); err != nil {
		t.Fatalf("creating nginx ingress resource: %s", err)
	}

	client, err := commonutil.GetClient()

	if err != nil {
		t.Fatal(errors.Wrap(err, "getting kubernetes client"))
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx-svc"}))
	if err := commonutil.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		t.Fatal(errors.Wrap(err, "waiting for nginx pods"))
	}

	if err := commonutil.WaitForService(client, "default", "nginx-svc", true, 1*time.Second, 2*time.Minute); err != nil {
		t.Fatal(errors.Wrap(err, "Error waiting for nginx service to be up"))
	}

	t.Log("getting nginx ingress...")

	nginxIP := ""

	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		cmd := []string{"get", "svc", "nginx-svc", "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}"}
		stdout, err := kubectlRunner.RunCommand(cmd)
		switch {
		case err == nil:
			nginxIP = string(stdout)
			return len(stdout) != 0, nil
		case !commonutil.IsRetryableAPIError(err):
			t.Errorf("`%s` failed with non retriable error: %v", cmd, err)
			return false, err
		default:
			t.Errorf("`%s` failed: %v", cmd, err)
			return false, nil
		}
	})

	if err != nil {
		t.Errorf("error getting ingress IP for nginx: %s", err)
	}

	if len(nginxIP) == 0 {
		stdout, err := kubectlRunner.RunCommand([]string{"get", "svc", "nginx-svc", "-o", "jsonpath={.status}"})

		if err != nil {
			t.Errorf("error debugging nginx service: %s", err)
		}

		t.Fatalf("svc should have ingress after tunnel is created, but it was empty! Result of `kubectl describe svc nginx-svc`:\n %s", string(stdout))
	}

	responseBody, err := getResponseBody(nginxIP)
	if err != nil {
		t.Fatalf("error reading from nginx at address(%s): %s", nginxIP, err)
	}
	if !strings.Contains(responseBody, "Welcome to nginx!") {
		t.Fatalf("response body doesn't seem like an nginx response:\n%s", responseBody)
	}
}

func getResponseBody(address string) (string, error) {
	httpClient := http.DefaultClient
	httpClient.Timeout = 5 * time.Second

	var resp *http.Response
	var err error

	request := func() error {
		resp, err = httpClient.Get(fmt.Sprintf("http://%s", address))
		if err != nil {
			retriable := &commonutil.RetriableError{Err: err}
			return retriable
		}
		return nil
	}

	if err = commonutil.RetryAfter(5, request, 1*time.Second); err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return "", errors.Wrapf(err, "error reading body, len bytes read: %d", len(body))
	}

	return string(body), nil
}
