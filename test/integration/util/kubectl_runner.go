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
	"math/rand"
	"os/exec"
	"testing"
	"time"

	commonutil "k8s.io/minikube/pkg/util"
)

const kubectlBinary = "kubectl"

// KubectlRunner runs a command using kubectl
type KubectlRunner struct {
	Profile    string // kube-context maps to a minikube profile
	T          *testing.T
	BinaryPath string
}

// NewKubectlRunner creates a new KubectlRunner
func NewKubectlRunner(t *testing.T, profile ...string) *KubectlRunner {
	if profile == nil {
		profile = []string{"minikube"}
	}

	p, err := exec.LookPath(kubectlBinary)
	if err != nil {
		t.Fatalf("Couldn't find kubectl on path.")
	}
	return &KubectlRunner{Profile: profile[0], BinaryPath: p, T: t}
}

// RunCommandParseOutput runs a command and parses the JSON output
func (k *KubectlRunner) RunCommandParseOutput(args []string, outputObj interface{}, setKubeContext ...bool) error {
	kubecContextArg := fmt.Sprintf(" --context=%s", k.Profile)
	args = append([]string{kubecContextArg}, args...) // prepending --context so it can be with with -- space

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
	kubecContextArg := fmt.Sprintf(" --context=%s", k.Profile)
	args = append([]string{kubecContextArg}, args...) // prepending --context so it can be with with -- space

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

func genRandString(strLen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UTC().UnixNano())
	result := make([]byte, strLen)
	for i := 0; i < strLen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// DeleteNamespace deletes the namespace
func (k *KubectlRunner) DeleteNamespace(namespace string) error {
	_, err := k.RunCommand([]string{"delete", "namespace", namespace})
	return err
}
