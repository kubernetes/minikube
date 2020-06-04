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
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
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
	start := time.Now()
	code := m.Run()
	fmt.Printf("Tests completed in %s (result code %d)\n", time.Since(start), code)
	os.Exit(code)
}

// StartArgs returns the arguments normally used for starting minikube
func StartArgs() []string {
	return strings.Split(*startArgs, " ")
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

// DockerDriver returns whether or not this test is using the docker or podman driver
func DockerDriver() bool {
	return strings.Contains(*startArgs, "--driver=docker") || strings.Contains(*startArgs, "--vm-driver=docker")
}

// PodmanDriver returns whether or not this test is using the docker or podman driver
func PodmanDriver() bool {
	return strings.Contains(*startArgs, "--vm-driver=podman") || strings.Contains(*startArgs, "driver=podman")
}

// KicDriver returns whether or not this test is using the docker or podman driver
func KicDriver() bool {
	return DockerDriver() || PodmanDriver()
}

// NeedsPortForward returns access to endpoints with this driver needs port forwarding
// (Docker on non-Linux platforms requires ports to be forwarded to 127.0.0.1)
func NeedsPortForward() bool {
	return KicDriver() && (runtime.GOOS == "windows" || runtime.GOOS == "darwin") || isMicrosoftWSL()
}

// isMicrosoftWSL will return true if process is running in WSL in windows
// checking for WSL env var based on this https://github.com/microsoft/WSL/issues/423#issuecomment-608237689
// also based on https://github.com/microsoft/vscode/blob/90a39ba0d49d75e9a4d7e62a6121ad946ecebc58/resources/win32/bin/code.sh#L24
func isMicrosoftWSL() bool {
	return os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSLPATH") != "" || os.Getenv("WSLENV") != ""
}

// CanCleanup returns if cleanup is allowed
func CanCleanup() bool {
	return *cleanup
}

// Minutes will return timeout in minutes based on how slow the machine is
func Minutes(n int) time.Duration {
	return time.Duration(*timeOutMultiplier) * time.Duration(n) * time.Minute
}

// Seconds will return timeout in minutes based on how slow the machine is
func Seconds(n int) time.Duration {
	return time.Duration(*timeOutMultiplier) * time.Duration(n) * time.Second
}
