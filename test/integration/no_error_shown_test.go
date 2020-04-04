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

// TestNoErrorShown asserts that there are no errors displayed
func TestNoErrorShown(t *testing.T) {
	if NoneDriver() {
		t.Skip("none driver always shows a warning")
	}
	MaybeParallel(t)

	if NeedsPortForward() {
		t.Skip("Docker for Desktop can be noisy")
	}

	profile := UniqueProfileName("no-error")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(20))
	defer CleanupWithLogs(t, profile, cancel)

	// TODO: make this multi-node once it's stable
	args := append([]string{"start", "-p", profile, "--memory=1900", "--wait=false"}, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	if rr.Stderr.Len() > 0 {
		t.Errorf("unexpected stderr: %v", rr.Stderr.String())
	}

	stdout := rr.Stdout.String()
	keywords := []string{"error", "fail", "warning", "conflict"}
	for _, keyword := range keywords {
		if strings.Contains(stdout, keyword) {
			t.Errorf("unexpected %q in stdout: %s", keyword, stdout)
		}
	}
}
