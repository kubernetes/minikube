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
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/test/integration/util"
)

// TestMain is the test main
func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var startTimeout = flag.Duration("timeout", 25*time.Minute, "max duration to wait for a full minikube start")
var binaryPath = flag.String("binary", "../../out/minikube", "path to minikube binary")
var globalArgs = flag.String("minikube-args", "", "Arguments to pass to minikube")
var startArgs = flag.String("minikube-start-args", "", "Arguments to pass to minikube start")
var mountArgs = flag.String("minikube-mount-args", "", "Arguments to pass to minikube mount")
var testdataDir = flag.String("testdata-dir", "testdata", "the directory relative to test/integration where the testdata lives")
var parallel = flag.Bool("parallel", true, "run the tests in parallel, set false for run sequentially")

// NewMinikubeRunner creates a new MinikubeRunner
func NewMinikubeRunner(t *testing.T, profile string, extraStartArgs ...string) util.MinikubeRunner {
	return util.MinikubeRunner{
		Profile:      profile,
		BinaryPath:   *binaryPath,
		StartArgs:    *startArgs + fmt.Sprintf(" --wait-timeout=%s ", *startTimeout/2) + strings.Join(extraStartArgs, " "), // adding timeout per component
		GlobalArgs:   *globalArgs,
		MountArgs:    *mountArgs,
		TimeOutStart: *startTimeout, // timeout for all start
		T:            t,
	}
}

// isTestNoneDriver checks if the current test is for none driver
func isTestNoneDriver(t *testing.T) bool {
	t.Helper()
	return strings.Contains(*startArgs, "--vm-driver=none")
}

// profileName chooses a profile name based on the test name
// to be used in minikube and kubecontext across that test
func profileName(t *testing.T) string {
	t.Helper()
	if isTestNoneDriver(t) {
		return "minikube"
	}
	p := strings.Split(t.Name(), "/")[0] // for i.e, TestFunctional/SSH returns TestFunctional
	if p == "TestFunctional" {
		return "minikube"
	}
	return p
}

// shouldRunInParallel deterimines if test should run in parallel or not
func shouldRunInParallel(t *testing.T) bool {
	t.Helper()
	if !*parallel {
		return false
	}
	if isTestNoneDriver(t) {
		return false
	}
	p := strings.Split(t.Name(), "/")[0] // for i.e, TestFunctional/SSH returns TestFunctional
	return p != "TestFunctional"         // gosimple lint: https://staticcheck.io/docs/checks#S1008
}
