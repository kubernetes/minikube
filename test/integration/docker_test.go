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
	"context"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
)

func TestDocker(t *testing.T) {
	if isTestNoneDriver() {
		t.Skip("skipping test as none driver does not bundle docker")
	}
	p := profile(t)
	t.Parallel()
	mk := NewMinikubeRunner(t, p, "--wait=false")
	defer mk.TearDown(t)

	// Start a timer for all remaining commands, to display failure output before a panic.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	if _, _, err := mk.RunWithContext(ctx, "delete"); err != nil {
		t.Logf("pre-delete failed (probably ok): %v", err)
	}

	stdout, stderr, err := mk.StartWithStds(13*time.Minute, "--docker-env=FOO=BAR", "--docker-env=BAZ=BAT", "--docker-opt=debug", " --docker-opt=icc=true")
	if err != nil {
		t.Fatalf("TestDocker minikube start failed : %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	mk.CheckStatus(state.Running.String())

	stdout, stderr, err = mk.RunWithContext(ctx, "ssh -- systemctl show docker --property=Environment --no-pager")
	if err != nil {
		t.Errorf("docker env: %v\nstderr: %s", err, stderr)
	}

	for _, envVar := range []string{"FOO=BAR", "BAZ=BAT"} {
		if !strings.Contains(stdout, envVar) {
			t.Errorf("Env var %s missing: %s.", envVar, stdout)
		}
	}

	stdout, stderr, err = mk.RunWithContext(ctx, "ssh -- systemctl show docker --property=ExecStart --no-pager")
	if err != nil {
		t.Errorf("ssh show docker: %v\nstderr: %s", err, stderr)
	}
	for _, opt := range []string{"--debug", "--icc=true"} {
		if !strings.Contains(stdout, opt) {
			t.Fatalf("Option %s missing from ExecStart: %s.", opt, stdout)
		}
	}
}
