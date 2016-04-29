// +build integration

/*
Copyright 2015 The Kubernetes Authors All rights reserved.
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
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath = flag.String("binary", "../minikube", "path to minikube binary")

func runCommand(t *testing.T, command string, checkError bool) string {
	path, _ := filepath.Abs(*binaryPath)
	cmd := exec.Command(path, command)
	stdout, err := cmd.Output()

	if checkError && err != nil {
		t.Fatalf("Error running command: %s %s", command, err)
	}
	return string(stdout)
}

func TestStartStop(t *testing.T) {

	getStatus := func() string {
		status := runCommand(t, "status", true)
		return strings.Trim(status, "\n")
	}

	checkStatus := func(desired string) {
		s := getStatus()
		if s != desired {
			t.Fatalf("Machine is in the wrong state: %s, expected  %s", s, desired)
		}
	}

	runCommand(t, "delete", false)
	checkStatus("Does Not Exist")

	runCommand(t, "start", true)
	checkStatus("Running")

	runCommand(t, "stop", true)
	checkStatus("Stopped")

	runCommand(t, "start", true)
	checkStatus("Running")

	runCommand(t, "delete", true)
	checkStatus("Does Not Exist")
}
