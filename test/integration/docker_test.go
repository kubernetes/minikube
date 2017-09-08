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
)

func TestDocker(t *testing.T) {
	minikubeRunner := NewMinikubeRunner(t)

	if strings.Contains(minikubeRunner.StartArgs, "--vm-driver=none") {
		t.Skip("skipping test as none driver does not bundle docker")
	}

	minikubeRunner.RunCommand("delete", false)
	startCmd := fmt.Sprintf("start %s %s %s", minikubeRunner.StartArgs, minikubeRunner.Args, "--docker-env=FOO=BAR --docker-env=BAZ=BAT --docker-opt=debug --docker-opt=icc=true")
	minikubeRunner.RunCommand(startCmd, true)
	minikubeRunner.EnsureRunning()

	dockerdEnvironment := minikubeRunner.RunCommand("ssh -- systemctl show docker --property=Environment --no-pager", true)
	fmt.Println(dockerdEnvironment)
	for _, envVar := range []string{"FOO=BAR", "BAZ=BAT"} {
		if !strings.Contains(dockerdEnvironment, envVar) {
			t.Fatalf("Env var %s missing from Environment: %s.", envVar, dockerdEnvironment)
		}
	}

	dockerdExecStart := minikubeRunner.RunCommand("ssh -- systemctl show docker --property=ExecStart --no-pager", true)
	fmt.Println(dockerdExecStart)
	for _, opt := range []string{"--debug", "--icc=true"} {
		if !strings.Contains(dockerdExecStart, opt) {
			t.Fatalf("Option %s missing from ExecStart: %s.", opt, dockerdExecStart)
		}
	}
}
