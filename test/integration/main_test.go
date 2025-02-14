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
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/detect"
)

// General configuration: used to set the VM Driver
var startArgs = flag.String("minikube-start-args", "", "Arguments to pass to minikube start")

// Flags for faster local integration testing
var forceProfile = flag.String("profile", "", "force tests to run against a particular profile")
var cleanup = flag.Bool("cleanup", true, "cleanup failed test run")
var enableGvisor = flag.Bool("gvisor", false, "run gvisor integration test (slow)")
var postMortemLogs = flag.Bool("postmortem-logs", true, "show logs after a failed test run")
var timeOutMultiplier = flag.Float64("timeout-multiplier", 1, "multiply the timeout for the tests")

// Paths to files - normally set for CI
var binaryPath = flag.String("binary", "../../out/minikube", "path to minikube binary")
var testdataDir = flag.String("testdata-dir", "testdata", "the directory relative to test/integration where the testdata lives")

// Node names are consistent, let's store these for easy access later
const (
	SecondNodeName = "m02"
	ThirdNodeName  = "m03"
)

// TestMain is the test main
func TestMain(m *testing.M) {
	flag.Parse()
	setMaxParallelism()

	start := time.Now()
	code := m.Run()
	fmt.Printf("Tests completed in %s (result code %d)\n", time.Since(start), code)
	os.Exit(code)
}

func TestMainNoArgs(t *testing.T) {
	rr, err := Run(t, exec.CommandContext(context.Background(), Target()))
	if err != nil {
		t.Fatalf("failed running minikube with no args %q: %v", rr.Command(), err)
	}
}

// setMaxParallelism caps the max parallelism. Go assumes 1 core per test, whereas minikube needs 2 cores per test.
func setMaxParallelism() {

	flagVal := flag.Lookup("test.parallel").Value.String()
	requested, err := strconv.Atoi(flagVal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to parse --test.parallel value: %q\n", flagVal)
		return
	}

	maxp := runtime.GOMAXPROCS(0)

	// Do not ignore what the user has explicitly set
	if requested != maxp {
		fmt.Fprintf(os.Stderr, "--test-parallel=%d was set via flags (system has %d cores)\n", requested, maxp)
		return
	}

	if maxp == 2 {
		fmt.Fprintf(os.Stderr, "Found %d cores, will not round down core count.\n", maxp)
		return
	}

	// Each "minikube start" consumes up to 2 cores, though the average usage is somewhat lower
	limit := int(math.Floor(float64(maxp) / 1.75))

	// Windows and MacOS tests were failing from timeouts due to too much parallelism
	if runtime.GOOS == "windows" {
		limit /= 2
	}

	// Hardcode limit to 2 for macOS
	if runtime.GOOS == "darwin" {
		limit = 2
	}

	fmt.Fprintf(os.Stderr, "Found %d cores, limiting parallelism with --test.parallel=%d\n", maxp, limit)
	if err := flag.Set("test.parallel", strconv.Itoa(limit)); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set test.parallel: %v\n", err)
	}
	runtime.GOMAXPROCS(limit)
}

// StartArgs returns the arguments normally used for starting minikube
func StartArgs() []string {
	return strings.Split(*startArgs, " ")
}

type ContextKey string

func StartArgsWithContext(ctx context.Context) []string {
	res := strings.Split(*startArgs, " ")
	value := ctx.Value(ContextKey("k8sVersion"))
	if value != nil && value != "" {
		res = append(res, fmt.Sprintf("--kubernetes-version=%s", value))
	}
	return res
}

// Target returns where the minikube binary can be found
func Target() string {
	return *binaryPath
}

// NoneDriver returns whether or not this test is using the none driver
func NoneDriver() bool {
	return strings.Contains(*startArgs, "--driver=none") || strings.Contains(*startArgs, "--vm-driver=none")
}

// HyperVDriver returns whether or not this test is using the Hyper-V driver
func HyperVDriver() bool {
	return strings.Contains(*startArgs, "--driver=hyperv") || strings.Contains(*startArgs, "--vm-driver=hyperv")
}

// VirtualboxDriver returns whether or not this test is using the VirtualBox driver
func VirtualboxDriver() bool {
	return strings.Contains(*startArgs, "--driver=virtualbox") || strings.Contains(*startArgs, "--vm-driver=virtualbox")
}

// DockerDriver returns whether or not this test is using the docker or podman driver
func DockerDriver() bool {
	return strings.Contains(*startArgs, "--driver=docker") || strings.Contains(*startArgs, "--vm-driver=docker")
}

// PodmanDriver returns whether or not this test is using the docker or podman driver
func PodmanDriver() bool {
	return strings.Contains(*startArgs, "--driver=podman") || strings.Contains(*startArgs, "--vm-driver=podman")
}

// RootlessDriver returns whether or not this test is using the rootless KIC driver
func RootlessDriver() bool {
	return strings.Contains(*startArgs, "--rootless")
}

// KicDriver returns whether or not this test is using the docker or podman driver
func KicDriver() bool {
	return DockerDriver() || PodmanDriver()
}

// VMDriver checks if the driver is a VM
func VMDriver() bool {
	return !KicDriver() && !NoneDriver()
}

// ContainerRuntime returns the name of a specific container runtime if it was specified
func ContainerRuntime() string {
	flag := "--container-runtime="
	for _, s := range StartArgs() {
		if strings.HasPrefix(s, flag) {
			return strings.TrimPrefix(s, flag)
		}
	}
	return constants.Docker
}

// arm64Platform returns true if running on arm64/* platform
func arm64Platform() bool {
	return runtime.GOARCH == "arm64"
}

// amd64Platform returns true if running on amd64/* platform
func amd64Platform() bool {
	return runtime.GOARCH == "amd64"
}

// NeedsPortForward returns access to endpoints with this driver needs port forwarding
// (Docker on non-Linux platforms and rootless KIC requires ports to be forwarded to 127.0.0.1)
func NeedsPortForward() bool {
	return KicDriver() && (runtime.GOOS == "windows" || runtime.GOOS == "darwin") || detect.IsMicrosoftWSL() || RootlessDriver()
}

// CanCleanup returns if cleanup is allowed
func CanCleanup() bool {
	return *cleanup
}

// Minutes will return timeout in minutes based on how slow the machine is
func Minutes(n int) time.Duration {
	return time.Duration(*timeOutMultiplier) * time.Duration(n) * time.Minute
}

// Seconds will return timeout in seconds based on how slow the machine is
func Seconds(n int) time.Duration {
	return time.Duration(*timeOutMultiplier) * time.Duration(n) * time.Second
}

// TestingKicBaseImage will return true if the integraiton test is running against a passed --base-image flag
func TestingKicBaseImage() bool {
	return strings.Contains(*startArgs, "base-image")
}
