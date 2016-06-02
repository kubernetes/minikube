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
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type MinikubeRunner struct {
	T          *testing.T
	BinaryPath string
}

func (m *MinikubeRunner) RunCommand(command string, checkError bool) string {
	commandArr := strings.Split(command, " ")
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, commandArr...)
	stdout, err := cmd.Output()

	if checkError && err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			m.T.Fatalf("Error running command: %s %s. Output: %s", command, exitError.Stderr, stdout)
		} else {
			m.T.Fatalf("Error running command: %s %s. Output: %s", command, err, stdout)
		}
	}
	return string(stdout)
}

func (m *MinikubeRunner) EnsureRunning() {
	if m.GetStatus() != "Running" {
		m.RunCommand("start", true)
	}
	m.CheckStatus("Running")
}

func (m *MinikubeRunner) SetEnvFromEnvCmdOutput(dockerEnvVars string) error {
	lines := strings.Split(dockerEnvVars, "\n")
	var envKey, envVal string
	seenEnvVar := false
	for _, line := range lines {
		if _, err := fmt.Sscanf(line, "export %s=%s", envKey, envVal); err != nil {
			seenEnvVar = true
			os.Setenv(envKey, envVal)
		}
	}
	if seenEnvVar == false {
		return fmt.Errorf("Error: No environment variables were found in docker-env command output: ", dockerEnvVars)
	}
	return nil
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

func (k *KubectlRunner) RunCommandParseOutput(args []string, outputObj interface{}) error {
	args = append(args, "-o=json")
	output, err := k.RunCommand(args)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(output))
	if err := d.Decode(outputObj); err != nil {
		return err
	}
	return nil
}

func (k *KubectlRunner) RunCommand(args []string) ([]byte, error) {
	cmd := exec.Command(k.BinaryPath, args...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return make([]byte, 0), fmt.Errorf("Error running command. Error  %s. Output: %s", err, stdout)
	}
	return stdout, nil
}

func (k *KubectlRunner) CreateRandomNamespace() string {
	const strLen = 20
	name := genRandString(strLen)
	k.RunCommand([]string{"create", "namespace", name})
	return name
}

func genRandString(strLen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UTC().UnixNano())
	result := make([]byte, strLen)
	for i := 0; i < strLen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func (k *KubectlRunner) DeleteNamespace(namespace string) error {
	_, err := k.RunCommand([]string{"delete", "namespace", namespace})
	return err
}
