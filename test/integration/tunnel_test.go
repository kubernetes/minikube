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
	"testing"
	"k8s.io/minikube/test/integration/util"
	"path/filepath"
	"fmt"
	"net/http"
	"strings"
	"time"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	commonutil "k8s.io/minikube/pkg/util"
	"io/ioutil"
)

func testTunnel(t *testing.T) {
	t.Log("starting tunnel...")
	runner := NewMinikubeRunner(t)
	go func() {
		output := runner.RunCommand("tunnel --alsologtostderr -v 8", true)
		fmt.Println(output)
	}()

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

	if err := commonutil.WaitForService(client, "default", "nginx-svc", true, time.Millisecond*500, time.Minute*10); err != nil {
		t.Errorf("Error waiting for nginx service to be up")
	}

	t.Log("getting nginx ingress...")

	nginxIP := ""


	for i := 1; i < 3 && len(nginxIP) == 0; i++ {
		stdout, err := kubectlRunner.RunCommand([]string{"get", "svc", "nginx-svc", "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}"})

		if err != nil {
			t.Fatalf("error listing nginx service: %s", err)
		}
		nginxIP = fmt.Sprintf("%s", stdout)
		time.Sleep(1 * time.Second)
	}

	if len(nginxIP) == 0 {
		t.Fatal("svc should have ingress after tunnel is created, but it was empty!")
	}

	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Second
	resp, e := httpClient.Get(fmt.Sprintf("http://%s", nginxIP))

	if e != nil {
		t.Fatalf("error reading from nginx at address(%s): %s", nginxIP, e)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if e != nil || len(body) == 0 {
		t.Fatalf("error reading body from nginx at address(%s): error: %s, len bytes read: %d", nginxIP, e, len(body))
	}

	responseBody := fmt.Sprintf("%s", body)
	if !strings.Contains(responseBody,"Welcome to nginx!") {
		t.Fatalf("response body doesn't seem like an nginx response:\n%s", responseBody)
	}
}
