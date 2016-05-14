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

package util

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type MinikubeRunner struct {
	T          *testing.T
	BinaryPath string
}

func (m *MinikubeRunner) RunCommand(command string, checkError bool) string {
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, command)
	stdout, err := cmd.Output()

	if checkError && err != nil {
		m.T.Fatalf("Error running command: %s %s. Output: %s", command, err, stdout)
	}
	return string(stdout)
}

func (m *MinikubeRunner) GetStatus() string {
	status := m.RunCommand("status", true)
	return strings.Trim(status, "\n")
}

func (m *MinikubeRunner) CheckStatus(desired string) {
	s := m.GetStatus()
	if s != desired {
		m.T.Fatalf("Machine is in the wrong state: %s, expected  %s", s, desired)
	}
}

type KubectlRunner struct {
	T          *testing.T
	BinaryPath string
}

func NewKubectlRunner(t *testing.T) *KubectlRunner {
	p, err := exec.LookPath("kubectl")
	if err != nil {
		t.Fatalf("Couldn't find kubectl on path.")
	}
	return &KubectlRunner{BinaryPath: p, T: t}
}

func (k *KubectlRunner) RunCommand(args []string, outputObj interface{}) {
	args = append(args, "-o=json")
	cmd := exec.Command(k.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		k.T.Fatalf("Error running command: %s %s. Error: %s, Output: %s", k.BinaryPath, args, err, output)
	}
	d := json.NewDecoder(bytes.NewReader(output))
	if err := d.Decode(outputObj); err != nil {
		k.T.Fatalf("Error parsing output: %s", err)
	}
}
