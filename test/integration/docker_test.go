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
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDocker(t *testing.T) {
	mk := NewMinikubeRunner(t)
	if strings.Contains(mk.StartArgs, "--vm-driver=none") {
		t.Skip("skipping test as none driver does not bundle docker")
	}

	// Start a timer for all remaining commands, to display failure output before a panic.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Pre-cleanup: this usually fails, because no instance is running.
	mk.RunWithContext(ctx, "delete")

	startCmd := fmt.Sprintf("start %s %s %s", mk.StartArgs, mk.Args,
		"--docker-env=FOO=BAR --docker-env=BAZ=BAT --docker-opt=debug --docker-opt=icc=true")
	out, err := mk.RunWithContext(ctx, startCmd)
	if err != nil {
		t.Fatalf("start: %v\nstart out: %s", err, out)
	}

	mk.EnsureRunning()

	out, err = mk.RunWithContext(ctx, "ssh -- systemctl show docker --property=Environment --no-pager")
	if err != nil {
		t.Errorf("docker env: %v\ndocker env out: %s", err, out)
	}

	for _, envVar := range []string{"FOO=BAR", "BAZ=BAT"} {
		if !strings.Contains(string(out), envVar) {
			t.Errorf("Env var %s missing: %s.", envVar, out)
		}
	}

	out, err = mk.RunWithContext(ctx, "ssh -- systemctl show docker --property=ExecStart --no-pager")
	if err != nil {
		t.Errorf("ssh show docker: %v\nshow docker out: %s", err, out)
	}
	for _, opt := range []string{"--debug", "--icc=true"} {
		if !strings.Contains(string(out), opt) {
			t.Fatalf("Option %s missing from ExecStart: %s.", opt, out)
		}
	}
}
