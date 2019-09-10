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

// the name of this file starts with z intentionally to make it run last after all other tests
// the intent is to make sure os env proxy settings be done after all other tests.
// for example in the case the test proxy clean up gets killed or fails
package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
)

// TestDashboardWithProxy starts a minikube cluster using a proxy, and asserts that dashboard still works
func TestDashboardWithProxy(t *testing.T) {
	srv, err := setUpProxy(t)
	if err != nil {
		t.Fatalf("Failed to set up the test proxy: %s", err)
	}

	profile := Profile("proxy")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer func() {
		CleanupWithLogs(t, profile, cancel)
		err := srv.Shutdown(context.Background())
		if err != nil {
			t.Errorf("Error shutting down the http proxy")
		}
	}()

	args := []string{"env", fmt.Sprintf("HTTP_PROXY=%s", srv.Addr), "NO_PROXY=", Target(), "start", "-p", profile, "--wait=false"}
	rr, err := Run(ctx, t, "/usr/bin/env", args...)
	if err != nil {
		t.Errorf("%s failed: %v", args, err)
	}

	want := "Found network options:"
	if !strings.Contains(rr.Stdout.String(), want) {
		t.Errorf("start stdout=%s, want: *%s*", rr.Stdout.String(), want)
	}

	want = "You appear to be using a proxy"
	if !strings.Contains(rr.Stderr.String(), want) {
		t.Errorf("start stderr=%s, want: *%s*", rr.Stderr.String(), want)
	}

	args = []string{"dashboard", "--url", "-p", profile, "--alsologtostderr", "-v=1"}
	ss, err := Start(ctx, t, Target(), args...)
	if err != nil {
		t.Errorf("%s failed: %v", args, err)
	}
	defer func() {
		ss.Stop(t)
	}()

	start := time.Now()
	s, err := ReadLineWithTimeout(ss.Stdout, 300*time.Second)
	if err != nil {
		t.Fatalf("failed to read url within %s: %v\n", time.Since(start), err)
	}

	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		t.Fatalf("failed to parse %q: %v", s, err)
	}

	resp, err := retryablehttp.Get(u.String())
	if err != nil {
		t.Errorf("failed get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unable to read http response body: %v", err)
		}
		t.Errorf("%s returned status code %d, expected %d.\nbody:\n%s", u, resp.StatusCode, http.StatusOK, body)
	}
}

// setUpProxy runs a local http proxy and sets the env vars for it.
func setUpProxy(t *testing.T) (*http.Server, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get an open port")
	}

	addr := fmt.Sprintf("localhost:%d", port)
	proxy := goproxy.NewProxyHttpServer()
	srv := &http.Server{Addr: addr, Handler: proxy}
	go func(s *http.Server, t *testing.T) {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("Failed to start http server for proxy mock")
		}
	}(srv, t)
	return srv, nil
}
