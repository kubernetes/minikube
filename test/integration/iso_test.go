// +build iso

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
	"fmt"
	"strings"
	"testing"

	"k8s.io/minikube/test/integration/util"
)

func TestISO(t *testing.T) {

	minikubeRunner := util.MinikubeRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t}

	minikubeRunner.RunCommand("delete", true)
	minikubeRunner.Start()

	t.Run("permissions", testMountPermissions)
}

func testMountPermissions(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t}
	// test mount permissions
	mountPoints := []string{"/Users", "/hosthome"}
	perms := "drwxr-xr-x"
	foundMount := false

	for _, dir := range mountPoints {
		output := minikubeRunner.SSH(fmt.Sprintf("ls -l %s", dir))
		if strings.Contains(output, "No such file or directory") {
			continue
		}
		foundMount = true
		if !strings.Contains(output, perms) {
			t.Fatalf("Incorrect permissions. Expected %s, got %s.", perms, output)
		}
	}
	if !foundMount {
		t.Fatalf("No shared mount found. Checked %s", mountPoints)
	}
}
