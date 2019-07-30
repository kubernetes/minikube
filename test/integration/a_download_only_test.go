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

// a_download_only_test.go filename starts with a, for the purpose that it runs before all parallel tests and downloads the images and caches them.
package integration

import (
	"testing"
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

	stdout, stderr, err := mk.Start("--download-only")
	if err != nil {
		t.Fatalf("%s minikube start failed : %v\nstdout: %s\nstderr: %s", p, err, stdout, stderr)
	}
	// TODO: add test to check if files are downloaded

}
