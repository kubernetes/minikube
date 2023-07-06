//go:build integration

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
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/constants"
)

// TestDockerFlags makes sure the --docker-env and --docker-opt parameters are respected
func TestDockerFlags(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support ssh or bundle docker")
	}
	if ContainerRuntime() != "docker" {
		t.Skipf("skipping: only runs with docker container runtime, currently testing %s", ContainerRuntime())
	}
	MaybeParallel(t)

	profile := UniqueProfileName("docker-flags")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	// Use the most verbose logging for the simplest test. If it fails, something is very wrong.
	args := append([]string{"start", "-p", profile, "--cache-images=false", "--memory=2048", "--install-addons=false", "--wait=false", "--docker-env=FOO=BAR", "--docker-env=BAZ=BAT", "--docker-opt=debug", "--docker-opt=icc=true", "--alsologtostderr", "-v=5"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "sudo systemctl show docker --property=Environment --no-pager"))
	if err != nil {
		t.Errorf("failed to 'systemctl show docker' inside minikube. args %q: %v", rr.Command(), err)
	}

	for _, envVar := range []string{"FOO=BAR", "BAZ=BAT"} {
		if !strings.Contains(rr.Stdout.String(), envVar) {
			t.Errorf("expected env key/value %q to be passed to minikube's docker and be included in: *%q*.", envVar, rr.Stdout)
		}
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "sudo systemctl show docker --property=ExecStart --no-pager"))
	if err != nil {
		t.Errorf("failed on the second 'systemctl show docker' inside minikube. args %q: %v", rr.Command(), err)
	}
	for _, opt := range []string{"--debug", "--icc=true"} {
		if !strings.Contains(rr.Stdout.String(), opt) {
			t.Fatalf("expected %q output to have include *%s* . output: %q", rr.Command(), opt, rr.Stdout)
		}
	}
}

// TestForceSystemdFlag tests the --force-systemd flag, as one would expect.
func TestForceSystemdFlag(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support ssh or bundle docker")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("force-systemd-flag")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	// Use the most verbose logging for the simplest test. If it fails, something is very wrong.
	args := append([]string{"start", "-p", profile, "--memory=2048", "--force-systemd", "--alsologtostderr", "-v=5"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	containerRuntime := ContainerRuntime()
	switch containerRuntime {
	case "docker":
		validateDockerSystemd(ctx, t, profile)
	case "containerd":
		validateContainerdSystemd(ctx, t, profile)
	case "crio":
		validateCrioSystemd(ctx, t, profile)
	}

}

// validateDockerSystemd makes sure the --force-systemd flag worked with the docker container runtime
func validateDockerSystemd(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "docker info --format {{.CgroupDriver}}"))
	if err != nil {
		t.Errorf("failed to get docker cgroup driver. args %q: %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Output(), "systemd") {
		t.Fatalf("expected systemd cgroup driver, got: %v", rr.Output())
	}
}

// validateContainerdSystemd makes sure the --force-systemd flag worked with the containerd container runtime
func validateContainerdSystemd(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "cat /etc/containerd/config.toml"))
	if err != nil {
		t.Errorf("failed to get containerd cgroup driver. args %q: %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Output(), "SystemdCgroup = true") {
		t.Fatalf("expected systemd cgroup driver, got: %v", rr.Output())
	}
}

// validateCrioSystemd makes sure the --force-systemd flag worked with the cri-o container runtime
func validateCrioSystemd(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "cat /etc/crio/crio.conf.d/02-crio.conf"))
	if err != nil {
		t.Errorf("failed to get cri-o cgroup driver. args %q: %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Output(), "cgroup_manager = \"systemd\"") {
		t.Fatalf("expected systemd cgroup driver, got: %v", rr.Output())
	}
}

// TestForceSystemdEnv makes sure the MINIKUBE_FORCE_SYSTEMD environment variable works just as well as the --force-systemd flag
func TestForceSystemdEnv(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support ssh or bundle docker")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("force-systemd-env")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--memory=2048", "--alsologtostderr", "-v=5"}, StartArgs()...)
	cmd := exec.CommandContext(ctx, Target(), args...)
	cmd.Env = append(os.Environ(), "MINIKUBE_FORCE_SYSTEMD=true")
	rr, err := Run(t, cmd)
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
	containerRuntime := ContainerRuntime()
	switch containerRuntime {
	case "docker":
		validateDockerSystemd(ctx, t, profile)
	case "containerd":
		validateContainerdSystemd(ctx, t, profile)
	}
}

// TestDockerEnvContainerd makes sure that minikube docker-env command works when the runtime is containerd
func TestDockerEnvContainerd(t *testing.T) {
	t.Log("running with", ContainerRuntime(), DockerDriver(), runtime.GOOS, runtime.GOARCH)
	if ContainerRuntime() != constants.Containerd || !DockerDriver() || runtime.GOOS != "linux" {
		t.Skip("skipping: TestDockerEnvContainerd can only be run with the containerd runtime on Docker driver")
	}
	profile := UniqueProfileName("dockerenv")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	// start the minikube with containerd runtime
	args := append([]string{"start", "-p", profile}, StartArgs()...)
	cmd := exec.CommandContext(ctx, Target(), args...)
	startResult, err := Run(t, cmd)
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", startResult.Command(), err)
	}
	time.Sleep(time.Second * 10)

	// execute 'minikube docker-env --ssh-host --ssh-add' and extract the 'DOCKER_HOST' environment value
	cmd = exec.CommandContext(ctx, "/bin/bash", "-c", fmt.Sprintf("%s docker-env --ssh-host --ssh-add -p %s", Target(), profile))
	result, err := Run(t, cmd)
	if err != nil {
		t.Errorf("failed to execute minikube docker-env --ssh-host --ssh-add, error: %v, output: %s", err, result.Output())
	}

	output := result.Output()
	groups := regexp.MustCompile(`DOCKER_HOST="(\S*)"`).FindStringSubmatch(output)
	if len(groups) < 2 {
		t.Errorf("DOCKER_HOST doesn't match expected format, output is %s", output)
	}
	dockerHost := groups[1]
	segments := strings.Split(dockerHost, ":")
	if len(segments) < 3 {
		t.Errorf("DOCKER_HOST doesn't match expected format, output is %s", dockerHost)
	}

	// get SSH_AUTH_SOCK
	groups = regexp.MustCompile(`SSH_AUTH_SOCK=(\S*)`).FindStringSubmatch(output)
	if len(groups) < 2 {
		t.Errorf("failed to acquire SSH_AUTH_SOCK, output is %s", output)
	}
	sshAuthSock := groups[1]
	// get SSH_AGENT_PID
	groups = regexp.MustCompile(`SSH_AGENT_PID=(\S*)`).FindStringSubmatch(output)
	if len(groups) < 2 {
		t.Errorf("failed to acquire SSH_AUTH_PID, output is %s", output)
	}
	sshAgentPid := groups[1]

	cmd = exec.CommandContext(ctx, "/bin/bash", "-c", fmt.Sprintf("SSH_AUTH_SOCK=%s SSH_AGENT_PID=%s DOCKER_HOST=%s docker version", sshAuthSock, sshAgentPid, dockerHost))

	result, err = Run(t, cmd)
	if err != nil {
		t.Fatalf("failed to execute 'docker version', error: %v, output: %s", err, result.Output())
	}
	// if we are really connecting to nerdctld inside node, docker version output should have word 'nerdctl'
	// If everything works properly, in the output of `docker version` you should be able to see something like
	/*
		Server:
			nerdctl:
			Version:          1.0.0
			buildctl:
			Version:          0.10.3
			GitCommit:        c8d25d9a103b70dc300a4fd55e7e576472284e31
			containerd:
			Version:          1.6.10
			GitCommit:        770bd0108c32f3fb5c73ae1264f7e503fe7b2661
	*/
	if !strings.Contains(result.Output(), "nerdctl") {
		t.Fatal("failed to detect keyword 'nerdctl' in output of docker version")
	}

	// now try to build an image
	cmd = exec.CommandContext(ctx, "/bin/bash", "-c", fmt.Sprintf("SSH_AUTH_SOCK=%s SSH_AGENT_PID=%s DOCKER_HOST=%s DOCKER_BUILDKIT=0 docker build -t local/minikube-dockerenv-containerd-test:latest testdata/docker-env", sshAuthSock, sshAgentPid, dockerHost))
	result, err = Run(t, cmd)
	if err != nil {
		t.Errorf("failed to build images, error: %v, output:%s", err, result.Output())
	}

	// and check whether that image is really available
	cmd = exec.CommandContext(ctx, "/bin/bash", "-c", fmt.Sprintf("SSH_AUTH_SOCK=%s SSH_AGENT_PID=%s DOCKER_HOST=%s docker image ls", sshAuthSock, sshAgentPid, dockerHost))
	result, err = Run(t, cmd)
	if err != nil {
		t.Fatalf("failed to execute 'docker image ls', error: %v, output: %s", err, result.Output())
	}
	if !strings.Contains(result.Output(), "local/minikube-dockerenv-containerd-test") {
		t.Fatal("failed to detect image 'local/minikube-dockerenv-containerd-test' in output of docker image ls")
	}
}
