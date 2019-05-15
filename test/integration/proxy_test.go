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
func setUpProxy(t *testing.T) error {
	port, err := freeport.GetFreePort()
	if err != nil {
		return errors.Wrap(err, "Failed to get an open port")
	}

	addr := fmt.Sprintf("localhost:%d", port)
	err = os.Setenv("NO_PROXY", "")
	if err != nil {
		return errors.Wrap(err, "Failed to set no proxy env")
	}
	err = os.Setenv("HTTP_PROXY", addr)
	if err != nil {
		return errors.Wrap(err, "Failed to set http proxy env")
	}

	proxy := goproxy.NewProxyHttpServer()
	go func() {
		err := http.ListenAndServe(addr, proxy)
		t.Fatalf("Failed to server a http server for proxy : %s ", err)
	}()
	return nil
}

func TestProxy(t *testing.T) {
	err := setUpProxy(t)
	if err != nil {
		t.Fatalf("Failed to set up the test proxy: %s", err)
	}
	t.Run("ConsoleWarnning", testProxyWarning)
	t.Run("DashboardProxy", testDashboard)

}

// testProxyWarning checks user is warned correctly about the proxy related env vars
func testProxyWarning(t *testing.T) {
	mk := NewMinikubeRunner(t)
	// Start a timer for all remaining commands, to display failure output before a panic.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	startCmd := fmt.Sprintf("start %s %s %s", mk.StartArgs, mk.Args,
		"--alsologtostderr --v=5")
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
