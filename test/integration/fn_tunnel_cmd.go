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
	"context"
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
	"k8s.io/client-go/kubernetes"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/tunnel"
	"k8s.io/minikube/pkg/util/retry"
)

func validateTunnelCmd(ctx context.Context, t *testing.T, client kubernetes.Interface, profile string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if runtime.GOOS != "windows" {
		// Otherwise minikube fails waiting for a password.
		if err := exec.Command("sudo", "-n", "route").Run(); err != nil {
			t.Skipf("password required to execute 'route', skipping testTunnel: %v", err)
		}
	}

	// Pre-Cleanup
	if err := tunnel.NewManager().CleanupNotRunningTunnels(); err != nil {
		t.Errorf("CleanupNotRunningTunnels: %v", err)
	}

	// Start the tunnel
	args := []string{"-p", profile, "tunnel", "--alsologtostderr", "-v=1"}
	ss, err := StartCmd(ctx, t, Target(), args...)
	defer func() {
		if err := ss.Stop(t); err != nil {
			t.Logf("Failed to kill mount: %v", err)
		}
	}()

	// Start the "nginx" pod.
	rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "testsvc.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx-svc"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		t.Fatal(errors.Wrap(err, "waiting for nginx pods"))
	}

	if err := kapi.WaitForService(client, "default", "nginx-svc", true, 1*time.Second, 2*time.Minute); err != nil {
		t.Fatal(errors.Wrap(err, "Error waiting for nginx service to be up"))
	}

	// Wait until the nginx-svc has a loadbalancer ingress IP
	nginxIP := ""
	err = wait.PollImmediate(1*time.Second, 3*time.Minute, func() (bool, error) {
		rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "get", "svc", "nginx-svc", "-o", "-f", "jsonpath={.status.loadBalancer.ingress[0].ip}")
		if err != nil {
			return false, err
		}
		if len(rr.Stdout.String()) > 0 {
			nginxIP = rr.Stdout.String()
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Errorf("nginx-svc svc.status.loadBalancer.ingress never got an IP")

		rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "get", "svc", "nginx-svc")
		if err != nil {
			t.Errorf("%s failed: %v", rr.Args, err)
		}
		t.Logf("kubectl get svc nginx-svc:\n%s", rr.Stdout)
	}

	// Try fetching against the IP
	var resp *http.Response

	req := func() error {
		h := &http.Client{Timeout: time.Second * 10}
		resp, err = h.Get(fmt.Sprintf("http://%s", nginxIP))
		if err != nil {
			retriable := &retry.RetriableError{Err: err}
			return retriable
		}
		defer resp.Body.Close()
		return nil
	}
	if err = retry.Expo(req, time.Millisecond*500, 2*time.Minute, 6); err != nil {
		t.Errorf("failed to contact nginx at %s: %v", nginxIP, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("ReadAll: %v", err)
	}
	want := "Welcome to nginx!"
	if !strings.Contains(string(body), want) {
		t.Errorf("body = %q, want *%s*", body, want)
	}
}
