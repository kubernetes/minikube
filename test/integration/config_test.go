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

func TestConfig(t *testing.T) {
	t.Parallel()
	p := profileName(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")

	tests := []struct {
		cmd    string
		stdout string
		stderr string
	}{
		{
			cmd:    "config unset cpus",
			stdout: "",
			stderr: "",
		},
		{
			cmd:    "config get cpus",
			stdout: "",
			stderr: "Error: specified key could not be found in config",
		},
		{
			cmd:    "config set cpus 2",
			stdout: "! These changes will take effect upon a minikube delete and then a minikube start",
			stderr: "",
		},
		{
			cmd:    "config get cpus",
			stdout: "2",
			stderr: "",
		},
		{
			cmd:    "config unset cpus",
			stdout: "",
			stderr: ""},
		{
			cmd:    "config get cpus",
			stdout: "",
			stderr: "Error: specified key could not be found in config",
		},
	}

	for _, tc := range tests {
		stdout, stderr := mk.RunCommand(tc.cmd, false)
		if !compare(tc.stdout, stdout) {
			t.Fatalf("Expected stdout to be: %s. Stdout was: %s Stderr: %s", tc.stdout, stdout, stderr)
		}
		if !compare(tc.stderr, stderr) {
			t.Fatalf("Expected stderr to be: %s. Stdout was: %s Stderr: %s", tc.stderr, stdout, stderr)
		}
	}
}

func compare(s1, s2 string) bool {
	return strings.TrimSpace(s1) == strings.TrimSpace(s2)
}
