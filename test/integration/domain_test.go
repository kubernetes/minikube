//go:build integration && linux

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
	"os/exec"
	"testing"
)

// Test to make sure the default domain works as intended
func TestDefaultDomainName(t *testing.T) {
	// MaybeParallel(t)

	profile := UniqueProfileName("domain-default")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := append([]string{Target(), "start", "-p", profile, "--wait=apiserver,system_pods,default_sa"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "run", "busybox", "--rm", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "/bin/sh", "-c", "awk '$1 == \"search\" {print $2; if ($2 !~ /\\.cluster\\.local$/) { print $2;exit 1}}' /etc/resolv.conf"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
		t.Errorf("Output: %s", rr.Output())
	}
}

// Test to make sure using dns-domain flag with no default-dns-domain config works as intended.
func TestDnsDomainFlag(t *testing.T) {
	// MaybeParallel(t)

	profile := UniqueProfileName("domain-flag")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := append([]string{Target(), "start", "-p", profile, "--dns-domain=flagcluster.example.com", "--wait=apiserver,system_pods,default_sa"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "run", "busybox", "--rm", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "/bin/sh", "-c", "awk '$1 == \"search\" {print $2; if ($2 !~ /\\.flagcluster\\.example\\.com$/) { print $2;exit 1}}' /etc/resolv.conf"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
		t.Errorf("Output: %s", rr.Output())
	}

	StopCluster(t, ctx, profile)

	// make sure the dns domain doesn't change when no flag is given on subsequent start
	startArgs = append([]string{Target(), "start", "-p", profile}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "run", "busybox", "--rm", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "/bin/sh", "-c", "awk '$1 == \"search\" {print $2; if ($2 !~ /\\.flagcluster\\.example\\.com$/) { print $2;exit 1}}' /etc/resolv.conf"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
		t.Errorf("Output: %s", rr.Output())
	}

	StopCluster(t, ctx, profile)

	// make sure the dns domain changes when a new flag is passed
	startArgs = append([]string{Target(), "start", "-p", profile, "--dns-domain=flagcluster2.example.com"}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "run", "busybox", "--rm", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "/bin/sh", "-c", "awk '$1 == \"search\" {print $2; if ($2 !~ /\\.flagcluster2\\.example\\.com$/) { print $2;exit 1}}' /etc/resolv.conf"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
		t.Errorf("Output: %s", rr.Output())
	}

}

// Test to make sure using default-dns-domain config with no dns-domain flag works as intended.
func TestDnsDomainConfig(t *testing.T) {
	// MaybeParallel(t)

	profile := UniqueProfileName("domain-config")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	defer CleanupWithLogs(t, profile, cancel)
	defer CleanupConfig(t, ctx, "default-dns-domain")

	SetConfig(t, ctx, "default-dns-domain", "configcluster.example.com")

	startArgs := append([]string{Target(), "start", "-p", profile, "--wait=apiserver,system_pods,default_sa"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "run", "busybox", "--rm", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "/bin/sh", "-c", "awk '$1 == \"search\" {print $2; if ($2 !~ /\\.configcluster\\.example\\.com$/) { print $2;exit 1}}' /etc/resolv.conf"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
		t.Errorf("Output: %s", rr.Output())
	}

	StopCluster(t, ctx, profile)

	// changing config doesn't change dns domain of restarted cluster
	SetConfig(t, ctx, "default-dns-domain", "configcluster2.example.com")

	startArgs = append([]string{Target(), "start", "-p", profile}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "run", "busybox", "--rm", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "/bin/sh", "-c", "awk '$1 == \"search\" {print $2; if ($2 !~ /\\.configcluster\\.example\\.com$/) { print $2;exit 1}}' /etc/resolv.conf"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
		t.Errorf("Output: %s", rr.Output())
	}

	StopCluster(t, ctx, profile)

	//should not change when removing the config
	CleanupConfig(t, ctx, "default-dns-domain")

	startArgs = append([]string{Target(), "start", "-p", profile}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "run", "busybox", "--rm", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "/bin/sh", "-c", "awk '$1 == \"search\" {print $2; if ($2 !~ /\\.configcluster\\.example\\.com$/) { print $2;exit 1}}' /etc/resolv.conf"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
		t.Errorf("Output: %s", rr.Output())
	}
}

// Test to make sure using dns-domain flash with default-dns-domain config works as intended.
func TestDnsDomainFlagAndDefaultDnsDomainConfig(t *testing.T) {
	profile := UniqueProfileName("domain-flag-and-config")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	defer CleanupWithLogs(t, profile, cancel)
	defer CleanupConfig(t, ctx, "default-dns-domain")

	SetConfig(t, ctx, "default-dns-domain", "configcluster.example.com")

	startArgs := append([]string{Target(), "start", "-p", profile, "--wait=apiserver,system_pods,default_sa", "--dns-domain=flagcluster.example.com"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "run", "busybox", "--rm", "--restart=Never", "--image=gcr.io/k8s-minikube/busybox", "-it", "--", "/bin/sh", "-c", "awk '$1 == \"search\" {print $2; if ($2 !~ /\\.flagcluster\\.example\\.com$/) { print $2;exit 1}}' /etc/resolv.conf"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
		t.Errorf("Output: %s", rr.Output())
	}
}
