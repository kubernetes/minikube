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

import (
	"strings"
	"testing"
)

// testProfileList tests the `minikube profile list` command
func testProfileList(t *testing.T) {
	t.Parallel()
	profile := "minikube"
	mk := NewMinikubeRunner(t, "--wait=false")
	out := mk.RunCommand("profile list", true)
	if !strings.Contains(out, profile) {
		t.Errorf("Error , failed to read profile name (%s) in `profile list` command output : \n %q ", profile, out)
	}
}
