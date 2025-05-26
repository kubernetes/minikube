//go:build integration

/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/util/retry"
)

// TestScheduledStopWindows tests the schedule stop functionality on Windows
func TestScheduledStopWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("test only runs on windows")
	}
	profile := UniqueProfileName("scheduled-stop")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(6))
	defer CleanupWithLogs(t, profile, cancel)
	startMinikube(ctx, t, profile)

	// schedule a stop for 5m from now
	stopMinikube(ctx, t, profile, []string{"--schedule", "5m"})
	// make sure timeToStop is present in status
	ensureMinikubeScheduledTime(ctx, t, profile, 5*time.Minute)
	// make sure the systemd service is running
	rr, err := Run(t, exec.CommandContext(ctx, Target(), []string{"ssh", "-p", profile, "--", "sudo", "systemctl", "show", constants.ScheduledStopSystemdService, "--no-page"}...))
	if err != nil {
		t.Fatalf("getting minikube-scheduled-stop status: %v\n%s", err, rr.Output())
	}
	if !strings.Contains(rr.Output(), "ActiveState=active") {
		t.Fatalf("minikube-scheduled-stop is not running: %v", rr.Output())
	}

	// reschedule stop for 5 seconds from now
	stopMinikube(ctx, t, profile, []string{"--schedule", "5s"})

	// wait for stop to complete
	time.Sleep(time.Minute)
	// make sure minikube timetoStop is not present
	ensureTimeToStopNotPresent(ctx, t, profile)
	// make sure minikube status is "Stopped"
	ensureMinikubeStatus(ctx, t, profile, "Host", state.Stopped.String())
}

// TestScheduledStopUnix tests the schedule stop functionality on Unix
func TestScheduledStopUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test only runs on unix")
	}
	if NoneDriver() {
		t.Skip("--schedule does not work with the none driver")
	}
	profile := UniqueProfileName("scheduled-stop")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer CleanupWithLogs(t, profile, cancel)
	startMinikube(ctx, t, profile)

	// schedule a stop for 5 min from now and make sure PID is created
	stopMinikube(ctx, t, profile, []string{"--schedule", "5m"})
	// make sure timeToStop is present in status
	ensureMinikubeScheduledTime(ctx, t, profile, 5*time.Minute)
	pid := checkPID(t, profile)
	if !processRunning(t, pid) {
		t.Fatalf("process %v is not running", pid)
	}

	// schedule a second stop which should cancel the first scheduled stop
	stopMinikube(ctx, t, profile, []string{"--schedule", "15s"})
	if processRunning(t, pid) {
		t.Fatalf("process %v running but should have been killed on reschedule of stop", pid)
	}
	pid = checkPID(t, profile)

	// cancel the shutdown and make sure minikube is still running after 8 seconds
	// sleep 12 just to be safe
	stopMinikube(ctx, t, profile, []string{"--cancel-scheduled"})
	time.Sleep(25 * time.Second)
	// make sure minikube status is "Running"
	ensureMinikubeStatus(ctx, t, profile, "Host", state.Running.String())
	// make sure minikube timetoStop is not present
	ensureTimeToStopNotPresent(ctx, t, profile)

	// schedule another stop, make sure minikube status is "Stopped"
	stopMinikube(ctx, t, profile, []string{"--schedule", "15s"})
	time.Sleep(15 * time.Second)
	if processRunning(t, pid) {
		t.Fatalf("process %v running but should have been killed on reschedule of stop", pid)
	}

	// wait for stop to complete
	time.Sleep(30 * time.Second)
	// make sure minikube timetoStop is not present
	ensureTimeToStopNotPresent(ctx, t, profile)
	// make sure minikube status is "Stopped"
	ensureMinikubeStatus(ctx, t, profile, "Host", state.Stopped.String())
}

func startMinikube(ctx context.Context, t *testing.T, profile string) {
	args := append([]string{"start", "-p", profile, "--memory=3072"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}
}

func stopMinikube(ctx context.Context, t *testing.T, profile string, additionalArgs []string) {
	args := []string{"stop", "-p", profile}
	args = append(args, additionalArgs...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("stopping minikube: %v\n%s", err, rr.Output())
	}
}

func checkPID(t *testing.T, profile string) string {
	file := localpath.PID(profile)
	var contents []byte
	getContents := func() error {
		var err error
		contents, err = os.ReadFile(file)
		return err
	}
	// first, make sure the PID file exists
	if err := retry.Expo(getContents, 100*time.Microsecond, time.Minute*1); err != nil {
		t.Fatalf("error reading %s: %v", file, err)
	}
	return string(contents)
}

func processRunning(t *testing.T, pid string) bool {
	// make sure PID file contains a running process
	p, err := strconv.Atoi(pid)
	if err != nil {
		return false
	}
	process, err := os.FindProcess(p)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	t.Log("signal error was: ", err)
	return err == nil
}
func ensureMinikubeStatus(ctx context.Context, t *testing.T, profile, key string, wantStatus string) {
	checkStatus := func() error {
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
		defer cancel()
		got := Status(ctx, t, Target(), profile, key, profile)
		if got != wantStatus {
			return fmt.Errorf("expected post-stop %q status to be -%q- but got *%q*", key, wantStatus, got)
		}
		return nil
	}
	if err := retry.Expo(checkStatus, time.Second, time.Minute); err != nil {
		t.Fatalf("error %v", err)
	}
}

func ensureMinikubeScheduledTime(ctx context.Context, t *testing.T, profile string, givenTime time.Duration) {
	checkTime := func() error {
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
		defer cancel()
		got := Status(ctx, t, Target(), profile, "TimeToStop", profile)
		gotTime, _ := time.ParseDuration(got)
		if gotTime < givenTime {
			return nil
		}
		return fmt.Errorf("expected scheduled stop TimeToStop to be less than *%q* but got *%q*", givenTime, got)
	}
	if err := retry.Expo(checkTime, time.Second, time.Minute); err != nil {
		t.Fatalf("error %v", err)
	}
}

func ensureTimeToStopNotPresent(ctx context.Context, t *testing.T, profile string) {
	args := []string{"status", "-p", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	// `minikube status` returns a non-zero exit code if the cluster is not running
	// so also check for "Error" in the output to confirm it's an actual error
	if err != nil && strings.Contains(rr.Output(), "Error") {
		t.Fatalf("minikube status: %v\n%s", err, rr.Output())
	}
	if strings.Contains(rr.Output(), "TimeToStop") {
		t.Fatalf("expected status output to not include `TimeToStop` but got *%s*", rr.Output())
	}
}
