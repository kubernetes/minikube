// +build integration

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

func TestDockerEnv(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t}

	minikubeRunner.RunCommand("delete", true)
	minikubeRunner.RunCommand("start --docker-env=FOO=BAR --docker-env=BAZ=BAT", true)
	minikubeRunner.EnsureRunning()

	profileContents := minikubeRunner.RunCommand("ssh cat /var/lib/boot2docker/profile", true)
	fmt.Println(profileContents)
	for _, envVar := range []string{"FOO=BAR", "BAZ=BAT"} {
		if !strings.Contains(profileContents, fmt.Sprintf("export \"%s\"", envVar)) {
			t.Fatalf("Env var %s missing from file: %s.", envVar, profileContents)
		}
	}
}
