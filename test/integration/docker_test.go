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
)

func TestDocker(t *testing.T) {
	profile := Profile("docker")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer Cleanup(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--wait=false", "--docker-env=FOO=BAR", "--docker-env=BAZ=BAT", "--docker-opt=debug", "--docker-opt=icc=true"}, StartArgs()...)
	rr, err := Run(ctx, t, Target(), args...)
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	rr, err = Run(ctx, t, Target(), "-p", profile, "ssh", "systemctl show docker --property=Environment --no-pager")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	for _, envVar := range []string{"FOO=BAR", "BAZ=BAT"} {
		if !strings.Contains(rr.Stdout.String(), envVar) {
			t.Errorf("env var %s missing: %s.", envVar, rr.Stdout)
		}
	}

	rr, err = Run(ctx, t, Target(), "-p", profile, "ssh", "systemctl show docker --property=RunStart --no-pager")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	for _, opt := range []string{"--debug", "--icc=true"} {
		if !strings.Contains(rr.Stdout.String(), opt) {
			t.Fatalf("Option %s missing from RunStart: %s.", opt, rr.Stdout)
		}
	}
}
