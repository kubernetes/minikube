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

	minikubeRunner.RunCommand("delete", false)

	startCmd := fmt.Sprintf("start %s %s", minikubeRunner.Args, "--docker-env=FOO=BAR --docker-env=BAZ=BAT")
	minikubeRunner.RunCommand(startCmd, true)
	minikubeRunner.EnsureRunning()

	filename := "/var/lib/boot2docker/profile"
	// Figure out if it's b2d or buildroot
	osContents := minikubeRunner.RunCommand("ssh cat /etc/os-release", true)
	t.Logf("os-release: %s", osContents)
	if strings.Contains(osContents, "NAME=Buildroot") {
		filename = "/etc/systemd/system/docker.service"
	}

	profileContents := minikubeRunner.RunCommand(fmt.Sprintf("ssh sudo cat %s", filename), true)
	fmt.Println(profileContents)
	for _, envVar := range []string{"FOO=BAR", "BAZ=BAT"} {
		if !strings.Contains(profileContents, envVar) {
			t.Fatalf("Env var %s missing from file: %s.", envVar, profileContents)
		}
	}
}
