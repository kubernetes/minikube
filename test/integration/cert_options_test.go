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
	"testing"
)

func TestCertOptions(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support ssh or bundle docker")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("cert-options")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	// Use the most verbose logging for the simplest test. If it fails, something is very wrong.
	args := append([]string{"start", "-p", profile, "--memory=1900", "--apiserver-ips=127.0.0.1,192.168.15.15", "--apiserver-names=localhost,www.google.com", "--apiserver-port=8555"}, StartArgs()...)

	// We can safely override --apiserver-name with
	if NeedsPortForward() {
		args = append(args, "--apiserver-name=localhost")
	}

	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	// test that file written from host was read in by the pod via cat /mount-9p/written-by-host;
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "version"))
	if err != nil {
		t.Errorf("failed to get kubectl version. args %q : %v", rr.Command(), err)
	}

}
