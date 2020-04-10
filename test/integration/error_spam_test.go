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
	"os/exec"
	"strings"
	"testing"
)

// TestErrorSpam asserts that there are no errors displayed
func TestErrorSpam(t *testing.T) {
	if NoneDriver() {
		t.Skip("none driver always shows a warning")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("nospam")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(25))
	defer CleanupWithLogs(t, profile, cancel)

	// This should likely use multi-node once it's ready
	args := append([]string{"start", "-p", profile, "-n=1", "--memory=2250", "--wait=false"}, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	for _, line := range strings.Split(rr.Stderr.String(), "\n") {
		if strings.HasPrefix(line, "E") {
			t.Errorf("unexpected error log in stderr: %q", line)
			continue
		}

		if strings.Contains(line, "kubectl") || strings.Contains(line, "slow") || strings.Contains(line, "long time") {
			continue
		}
		if len(strings.TrimSpace(line)) > 0 {
			t.Errorf("unexpected stderr line: %q", line)
		}
	}

	for _, line := range strings.Split(rr.Stdout.String(), "\n") {
		keywords := []string{"error", "fail", "warning", "conflict"}
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				t.Errorf("unexpected %q in stdout line: %q", keyword, line)
			}
		}
	}
}
