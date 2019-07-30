// +build integration

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

// this file name has to start with a so it be run before all other tests
import (
	"testing"
	"time"
)

// TestDownloadOnly downloads ISOs also tests the --download-only option
// Note this test runs before all tests (because of file name) and caches images for them
func TestDownloadOnly(t *testing.T) {
	p := profile(t)
	if isTestNoneDriver() {
		t.Skip()

	}
	mk := NewMinikubeRunner(t, p)
	if !isTestNoneDriver() { // none driver doesnt need to be deleted
		defer mk.TearDown(t)
	}

	stdout, stderr, err := mk.StartWithStds(15*time.Minute, "--download-only")
	if err != nil {
		t.Fatalf("%s minikube start failed : %v\nstdout: %s\nstderr: %s", p, err, stdout, stderr)
	}
	// TODO: add test to check if files are downloaded

}
