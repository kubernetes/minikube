//go:build integration

/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

	"k8s.io/minikube/pkg/minikube/constants"
)

// TestImageBuild makes sure the 'minikube image build' command works fine
func TestImageBuild(t *testing.T) {
	if ContainerRuntime() != constants.Docker {
		t.Skip()
	}
	type validateFunc func(context.Context, *testing.T, string)
	profile := UniqueProfileName("image")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(15))
	defer Cleanup(t, profile, cancel)

	// Serial tests
	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"Setup", validateSetupImageBuild},
			{"NormalBuild", validateNormalImageBuild},
			{"BuildWithBuildArg", validateImageBuildWithBuildArg},
			{"BuildWithDockerIgnore", validateImageBuildWithDockerIgnore},
			{"BuildWithSpecifiedDockerfile", validateNormalImageBuildWithSpecifiedDockerfile},
			{"validateImageBuildWithBuildEnv", validateImageBuildWithBuildEnv},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
			})
			// if setup fails bail
			if tc.name == "Setup" && t.Failed() {
				return
			}
		}
	})
}

// validateSetupImageBuild starts a cluster for the image builds
func validateSetupImageBuild(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	startArgs := append([]string{"start", "-p", profile}, StartArgs()...)
	if rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...)); err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
}

// validateNormalImageBuild is normal test case for minikube image build, with -t parameter
func validateNormalImageBuild(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	args := []string{"image", "build", "-t", "aaa:latest", "./testdata/image-build/test-normal", "-p", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to build image with args: %q : %v", rr.Command(), err)
	}
}

// validateNormalImageBuildWithSpecifiedDockerfile is normal test case for minikube image build, with -t and -f parameter
func validateNormalImageBuildWithSpecifiedDockerfile(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	args := []string{"image", "build", "-t", "aaa:latest", "-f", "inner/Dockerfile", "./testdata/image-build/test-f", "-p", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to build image with args: %q : %v", rr.Command(), err)
	}
}

// validateImageBuildWithBuildArg is a test case building with --build-opt
func validateImageBuildWithBuildArg(ctx context.Context, t *testing.T, profile string) {
	// test case for bug in https://github.com/kubernetes/minikube/issues/12384
	defer PostMortemLogs(t, profile)
	args := []string{"image", "build", "-t", "aaa:latest", "--build-opt=build-arg=ENV_A=test_env_str", "--build-opt=no-cache", "./testdata/image-build/test-arg", "-p", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to build image with args: %q : %v", rr.Command(), err)
	}
	output := rr.Output()
	if !strings.Contains(output, "test_env_str") {
		t.Fatalf("failed to pass build-args with args: %q : %s", rr.Command(), output)
	}
}

// validateImageBuildWithBuildEnv is a test case building with --build-env
func validateImageBuildWithBuildEnv(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// current this test cannot be passed because issue https://github.com/kubernetes/minikube/issues/12431 hasn't been fixed, so this test case is not enabled
	t.Skip("skipping due to https://github.com/kubernetes/minikube/issues/12431")

	args := []string{"image", "build", "--build-opt=help", "--build-env=DOCKER_BUILDKIT=1", ".", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to build image with args: %q : %v", rr.Command(), err)
	}
	output := rr.Stdout.String()
	if strings.Contains(output, "--cgroup-parent") {
		// when DOCKER_BUILDKIT=0, "docker build" has the --cgroup-parent parameter, and when DOCKER_BUILDKIT=1, DOCKER_BUILDKIT doesn't provide this option at all.
		// we have set --build-env=DOCKER_BUILDKIT=1 so --cgroup-parent should not appear in help page
		t.Fatalf("failed to pass envs for command %q : %v", rr.Command(), err)
	}
}

// validateImageBuildWithDockerIgnore is a test case building with .dockerignore
func validateImageBuildWithDockerIgnore(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	args := []string{"image", "build", "-t", "aaa:latest", "./testdata/image-build/test-normal", "--build-opt=no-cache", "-p", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to build image with args: %q : %v", rr.Command(), err)
	}
}
