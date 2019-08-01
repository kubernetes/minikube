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

var startTimeout = flag.Int("timeout", 25, "number of minutes to wait for minikube start")
var binaryPath = flag.String("binary", "../../out/minikube", "path to minikube binary")
var globalArgs = flag.String("minikube-args", "", "Arguments to pass to minikube")
var startArgs = flag.String("minikube-start-args", "", "Arguments to pass to minikube start")
var mountArgs = flag.String("minikube-mount-args", "", "Arguments to pass to minikube mount")
var testdataDir = flag.String("testdata-dir", "testdata", "the directory relative to test/integration where the testdata lives")
var disableParallel = flag.Bool("disable-parallel", false, "run the tests squentially and disable all parallel runs")

// NewMinikubeRunner creates a new MinikubeRunner
func NewMinikubeRunner(t *testing.T, profile string, extraStartArgs ...string) util.MinikubeRunner {
	return util.MinikubeRunner{
		Profile:      profile,
		BinaryPath:   *binaryPath,
		StartArgs:    *startArgs + " " + strings.Join(extraStartArgs, " "),
		GlobalArgs:   *globalArgs,
		MountArgs:    *mountArgs,
		TimeOutStart: time.Duration(*startTimeout) * time.Minute,
		T:            t,
	}
}

// isTestNoneDriver checks if the current test is for none driver
func isTestNoneDriver() bool {
	return strings.Contains(*startArgs, "--vm-driver=none")
}

// profileName chooses a profile name based on the test name
// to be used in minikube and kubecontext across that test
func profileName(t *testing.T) string {
	if isTestNoneDriver() {
		return "minikube"
	}
	p := t.Name()
	if strings.Contains(p, "/") { // for i.e, TestFunctional/SSH returns TestFunctional
		p = strings.Split(p, "/")[0]
	}
	if p == "TestFunctional" {
		return "minikube"
	}
	return p
}

// toParallel deterimines if test should run in  parallel or not
func toParallel() bool {
	if *disableParallel {
		return false
	}
	if isTestNoneDriver() {
		return false
	}
	return true
}
