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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDockerFlags(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support ssh or bundle docker")
	}
	MaybeParallel(t)
	profile := UniqueProfileName("docker-flags")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	tempDirPath, err := ioutil.TempDir("", profile)
	if err != nil {
		t.Errorf("Fail to create temporary dir: " + err.Error())
	}

	defer func() {
		err := os.RemoveAll(tempDirPath)
		if err != nil {
			t.Errorf("Fail to remove temporary dir: " + err.Error())
		}
	}()
	driverMount := fmt.Sprintf("%s:%s", tempDirPath, tempDirPath)
	defer CleanupWithLogs(t, profile, cancel)

	// Use the most verbose logging for the simplest test. If it fails, something is very wrong.
	args := append([]string{"start", "-p", profile, "--cache-images=false", "--memory=1800", "--install-addons=false", "--wait=false", "--docker-env=FOO=BAR", "--docker-env=BAZ=BAT", "--docker-opt=debug", "--docker-opt=icc=true", "--driver-mounts", driverMount, "--alsologtostderr", "-v=5"}, StartArgs()...)
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

	args = []string{"inspect", profile, "--format", "{{.HostConfig.Binds}}"}
	rr, err = Run(t, exec.CommandContext(ctx, "docker", args...))
	if err != nil {
		t.Error("Fail to inspect docker container: " + err.Error())
	}
	if !strings.Contains(rr.Stdout.String(), driverMount) {
		t.Error("Fail to check docker mounts: " + rr.Stdout.String())
	}
}

func TestForceSystemdFlag(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support ssh or bundle docker")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("force-systemd-flag")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	// Use the most verbose logging for the simplest test. If it fails, something is very wrong.
	args := append([]string{"start", "-p", profile, "--memory=1800", "--force-systemd", "--alsologtostderr", "-v=5"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "docker info --format {{.CgroupDriver}}"))
	if err != nil {
		t.Errorf("failed to get docker cgroup driver. args %q: %v", rr.Command(), err)
	}

	if !strings.Contains(rr.Output(), "systemd") {
		t.Fatalf("expected systemd cgroup driver, got: %v", rr.Output())
	}
}

func TestForceSystemdEnv(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support ssh or bundle docker")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("force-systemd-env")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--memory=1800", "--alsologtostderr", "-v=5"}, StartArgs()...)
	cmd := exec.CommandContext(ctx, Target(), args...)
	cmd.Env = append(os.Environ(), "MINIKUBE_FORCE_SYSTEMD=true")
	rr, err := Run(t, cmd)
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "docker info --format {{.CgroupDriver}}"))
	if err != nil {
		t.Errorf("failed to get docker cgroup driver. args %q: %v", rr.Command(), err)
	}

	if !strings.Contains(rr.Output(), "systemd") {
		t.Fatalf("expected systemd cgroup driver, got: %v", rr.Output())
	}
}
