// +build integration

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"os"
	"strings"
	"testing"
	"time"

	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
)

// setUpProxy runs a local http proxy and sets the env vars for it.
func setUpProxy(t *testing.T) (*http.Server, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get an open port")
	}

	addr := fmt.Sprintf("localhost:%d", port)
	err = os.Setenv("NO_PROXY", "")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to set no proxy env")
	}
	err = os.Setenv("HTTP_PROXY", addr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to set http proxy env")
	}

	proxy := goproxy.NewProxyHttpServer()
	srv := &http.Server{Addr: addr, Handler: proxy}
	go func(s *http.Server, t *testing.T) {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("Failed to start http server for proxy mock")
		}
	}(srv, t)
	return srv, nil
}

func TestProxy(t *testing.T) {
	origHP := os.Getenv("HTTP_PROXY")
	origNP := os.Getenv("NO_PROXY")
	srv, err := setUpProxy(t)
	if err != nil {
		t.Fatalf("Failed to set up the test proxy: %s", err)
	}

	defer func(t *testing.T) { // Clean up after setting up proxy
		err = os.Setenv("HTTP_PROXY", origHP)
		if err != nil {
			t.Errorf("Error reverting the HTTP_PROXY env")
		}
		err = os.Setenv("NO_PROXY", origNP)
		if err != nil {
			t.Errorf("Error reverting the NO_PROXY env")
		}

		err := srv.Shutdown(context.TODO()) // shutting down the http proxy after tests
		if err != nil {
			t.Errorf("Error shutting down the http proxy")
		}
	}(t)

	t.Run("ConsoleWarnning", testProxyWarning)
	t.Run("DashboardProxy", testDashboard)

}

// testProxyWarning checks user is warned correctly about the proxy related env vars
func testProxyWarning(t *testing.T) {
	mk := NewMinikubeRunner(t)
	// Start a timer for all remaining commands, to display failure output before a panic.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	startCmd := fmt.Sprintf("start %s %s %s", mk.StartArgs, mk.Args, "--alsologtostderr --v=5")
	stdout, stderr, err := mk.RunWithContext(ctx, startCmd)
	if err != nil {
		t.Fatalf("start: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	mk.EnsureRunning()

	// Pre-cleanup: this usually fails, because no instance is running.
	// mk.RunWithContext(ctx, "delete")
	msg := "Found network options:"
	if !strings.Contains(stdout, msg) {
		t.Errorf("Proxy wranning (%s) is missing from the output: %s", msg, stderr)
	}

	msg = "You appear to be using a proxy"
	if !strings.Contains(stderr, msg) {
		t.Errorf("Proxy wranning (%s) is missing from the output: %s", msg, stderr)
	}

}
