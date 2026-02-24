//go:build integration

/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

func TestDNSSearchFlag(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: dns-search is unsupported with none driver")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("dns-search")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := append([]string{
		"start",
		"-p", profile,
		"--memory=3072",
		"--wait=apiserver,system_pods,default_sa",
		"--dns-search=corp.example.com",
		"--dns-search=eng.example.com",
	}, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start minikube with dns-search. args %q: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "--", "cat /etc/resolv.conf"))
	if err != nil {
		t.Fatalf("failed to inspect node resolv.conf. args %q: %v", rr.Command(), err)
	}
	wantSearchLine := "search corp.example.com eng.example.com"
	if !strings.Contains(rr.Stdout.String(), wantSearchLine) {
		t.Errorf("node /etc/resolv.conf does not contain %q. output: %s", wantSearchLine, rr.Output())
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile,
		"run", "dns-search-resolv",
		"--restart=Never",
		"--image=gcr.io/k8s-minikube/busybox",
		"--labels=app=dns-search-resolv",
		"--command", "--", "sh", "-c", "cat /etc/resolv.conf"))
	if err != nil {
		t.Fatalf("failed to create pod to inspect pod resolv.conf. args %q: %v", rr.Command(), err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "app=dns-search-resolv", Minutes(2)); err != nil {
		t.Fatalf("pod did not become ready/succeeded in time: %v", err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "logs", "dns-search-resolv"))
	if err != nil {
		t.Fatalf("failed to read pod resolv.conf output. args %q: %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Stdout.String(), "corp.example.com") || !strings.Contains(rr.Stdout.String(), "eng.example.com") {
		t.Errorf("pod /etc/resolv.conf does not contain expected dns-search domains. output: %s", rr.Output())
	}
}
