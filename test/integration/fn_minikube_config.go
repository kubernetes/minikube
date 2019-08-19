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
	"testing"
)

func testMinikubeConfig(t *testing.T) {
	t.Parallel()
	p := profileName(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")

	tests := []struct {
		cmd    string
		output string
	}{
		{"config get memory", "Error: specified key could not be found in config"},
		{"config set memory 4016", "! These changes will take effect upon a minikube delete and then a minikube start"},
		{"config get memory", "4016"},
		{"config unset memory", ""},
		{"config get memory", "Error: specified key could not be found in config"},
	}

	for _, test := range tests {
		sshCmdOutput, stderr := mk.RunCommand(t.cmd)
		if !strings.Contains(sshCmdOutput, t.output) {
			t.Fatalf("ExpectedStr sshCmdOutput to be: %s. Output was: %s Stderr: %s", expectedStr, sshCmdOutput, stderr)
		}
	}
}
