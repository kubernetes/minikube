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
	"testing"

	"k8s.io/minikube/test/integration/util"
)

// TestMain is the test main
func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var binaryPath = flag.String("binary", "../../out/minikube", "path to minikube binary")
var args = flag.String("minikube-args", "", "Arguments to pass to minikube")
var startArgs = flag.String("minikube-start-args", "", "Arguments to pass to minikube start")
var mountArgs = flag.String("minikube-mount-args", "", "Arguments to pass to minikube mount")
var testdataDir = flag.String("testdata-dir", "testdata", "the directory relative to test/integration where the testdata lives")

// NewMinikubeRunner creates a new MinikubeRunner
func NewMinikubeRunner(t *testing.T) util.MinikubeRunner {
	return util.MinikubeRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		StartArgs:  *startArgs,
		MountArgs:  *mountArgs,
		T:          t,
	}
}
