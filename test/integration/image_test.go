// +build integration

/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"path/filepath"
	"testing"
)

// TestImage tests minikube image
func TestImage(t *testing.T) {
	profile := UniqueProfileName("image")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--memory=2048"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}

	if ContainerRuntime() == "containerd" {
		// sudo systemctl start buildkit.socket
		cmd := exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "--", "nohup",
			"sudo", "-b", "buildkitd", "--oci-worker=false",
			"--containerd-worker=true", "--containerd-worker-namespace=k8s.io")
		if rr, err = Run(t, cmd); err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}
		// unix:///run/buildkit/buildkitd.sock
	}

	tests := []struct {
		command string
		args    []string
	}{
		{
			command: "image",
			args:    []string{"load", filepath.Join(*testdataDir, "hello-world.tar")},
		}, {
			command: "image",
			args:    []string{"rm", "hello-world"},
		}, {
			command: "image",
			args:    []string{"build", "-t", "my-image", filepath.Join(*testdataDir, "build")},
		}, {
			command: "image",
			args:    []string{"list"},
		},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			args := []string{test.command, "-p", profile}
			args = append(args, test.args...)

			rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
			if err != nil {
				t.Errorf("failed to clean up: args %q: %v", rr.Command(), err)
			}
			if rr.Stdout.Len() > 0 {
				t.Logf("(dbg) Stdout: %s:\n%s", rr.Command(), rr.Stdout)
			}
			if rr.Stderr.Len() > 0 {
				t.Logf("(dbg) Stderr: %s:\n%s", rr.Command(), rr.Stderr)
			}
		})
	}
}
