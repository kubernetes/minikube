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
	"net"
	"net/url"
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
		t.Fatalf("Could not get kubernetes client: %s", err)
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"component": "kube-addon-manager"}))
	if err := pkgutil.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		t.Errorf("Error waiting for addon manager to be up")
	}
}

func testDashboard(t *testing.T) {
	t.Parallel()
	minikubeRunner := NewMinikubeRunner(t)

	if err := util.WaitForDashboardRunning(t); err != nil {
		t.Fatalf("waiting for dashboard to be up: %s", err)
	}

	dashboardURL := minikubeRunner.RunCommand("dashboard --url", true)
	u, err := url.Parse(strings.TrimSpace(dashboardURL))
	if err != nil {
		t.Fatalf("failed to parse dashboard URL %s: %v", dashboardURL, err)
	}
	if u.Scheme != "http" {
		t.Fatalf("wrong scheme in dashboard URL, expected http, actual %s", u.Scheme)
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("failed to split dashboard host %s: %v", u.Host, err)
	}
	if port != "30000" {
		t.Fatalf("Dashboard is exposed on wrong port, expected 30000, actual %s", port)
	}
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
