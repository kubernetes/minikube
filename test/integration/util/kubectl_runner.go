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

package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	commonutil "k8s.io/minikube/pkg/util"
)

const kubectlBinary = "kubectl"

// KubectlRunner runs a command using kubectl
type KubectlRunner struct {
	T          *testing.T
	BinaryPath string
}

// NewKubectlRunner creates a new KubectlRunner
func NewKubectlRunner(t *testing.T) *KubectlRunner {
	p, err := exec.LookPath(kubectlBinary)
	if err != nil {
		t.Fatalf("Couldn't find kubectl on path.")
	}
	return &KubectlRunner{BinaryPath: p, T: t}
}

// RunCommandParseOutput runs a command and parses the JSON output
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

// RunCommand runs a command, returning stdout
func (k *KubectlRunner) RunCommand(args []string) (stdout []byte, err error) {
	inner := func() error {
		cmd := exec.Command(k.BinaryPath, args...)
		stdout, err = cmd.CombinedOutput()
		if err != nil {
			retriable := &commonutil.RetriableError{Err: fmt.Errorf("error running command %s: %v. Stdout: \n %s", args, err, stdout)}
			k.T.Log(retriable)
			return retriable
		}
		return nil
	}

	err = commonutil.RetryAfter(3, inner, 2*time.Second)
	return stdout, err
}

// CreateRandomNamespace creates a random namespace
func (k *KubectlRunner) CreateRandomNamespace() string {
	const strLen = 20
	name := genRandString(strLen)
	if _, err := k.RunCommand([]string{"create", "namespace", name}); err != nil {
		k.T.Fatalf("Error creating namespace: %v", err)
	}
	return name
}

// CurrentContext returns the current context or fails
func (k *KubectlRunner) CurrentContext() string {
	out, err := k.RunCommand([]string{"config", "current-context"})
	ctx := strings.TrimRight(string(out), "\n")
	if err != nil {
		if strings.Contains(ctx, "is not set") {
			return ""
		}
		k.T.Fatalf("Failed to fetch current-context")
	}
	return ctx
}
