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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/minikube/pkg/minikube/assets"
	commonutil "k8s.io/minikube/pkg/util"
)

const kubectlBinary = "kubectl"

type MinikubeRunner struct {
	T          *testing.T
	BinaryPath string
	Args       string
	StartArgs  string
}

func (m *MinikubeRunner) Run(cmd string) error {
	_, err := m.SSH(cmd)
	return err
}

func (m *MinikubeRunner) Copy(f assets.CopyableFile) error {
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command("/bin/bash", "-c", path, "ssh", "--", fmt.Sprintf("cat >> %s", filepath.Join(f.GetTargetDir(), f.GetTargetName())))
	return cmd.Run()
}

func (m *MinikubeRunner) CombinedOutput(cmd string) (string, error) {
	return m.SSH(cmd)
}

func (m *MinikubeRunner) Remove(f assets.CopyableFile) error {
	_, err := m.SSH(fmt.Sprintf("rm -rf %s", filepath.Join(f.GetTargetDir(), f.GetTargetName())))
	return err
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

func (m *MinikubeRunner) RunDaemon(command string) *exec.Cmd {
	commandArr := strings.Split(command, " ")
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, commandArr...)
	err := cmd.Start()
	if err != nil {
		m.T.Fatalf("Error running command: %s %s", command, err)
	}
	return cmd
}

func (m *MinikubeRunner) SSH(command string) (string, error) {
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, "ssh", command)
	stdout, err := cmd.CombinedOutput()
	if err, ok := err.(*exec.ExitError); ok {
		return string(stdout), err
	}

	return string(stdout), nil
}

func (m *MinikubeRunner) Start() {
	m.RunCommand(fmt.Sprintf("start %s %s", m.StartArgs, m.Args), true)
}

func (m *MinikubeRunner) EnsureRunning() {
	if m.GetStatus() != "Running" {
		m.Start()
	}
	m.CheckStatus("Running")
}

func (m *MinikubeRunner) SetEnvFromEnvCmdOutput(dockerEnvVars string) error {
	re := regexp.MustCompile(`(\w+?) ?= ?"?(.+?)"?\n`)
	matches := re.FindAllStringSubmatch(dockerEnvVars, -1)
	seenEnvVar := false
	for _, m := range matches {
		seenEnvVar = true
		key, val := m[1], m[2]
		os.Setenv(key, val)
	}
	if !seenEnvVar {
		return fmt.Errorf("Error: No environment variables were found in docker-env command output: %s", dockerEnvVars)
	}
	return nil
}

func (m *MinikubeRunner) GetStatus() string {
	return m.RunCommand(fmt.Sprintf("status --format={{.MinikubeStatus}} %s", m.Args), false)
}

func (m *MinikubeRunner) GetLogs() string {
	return m.RunCommand(fmt.Sprintf("logs %s", m.Args), true)
}

func (m *MinikubeRunner) CheckStatus(desired string) {
	if err := m.CheckStatusNoFail(desired); err != nil {
		m.T.Fatalf("%v", err)
	}
}

func (m *MinikubeRunner) CheckStatusNoFail(desired string) error {
	s := m.GetStatus()
	if s != desired {
		return fmt.Errorf("Machine is in the wrong state: %s, expected  %s", s, desired)
	}
	return nil
}

type KubectlRunner struct {
	T          *testing.T
	BinaryPath string
}

func NewKubectlRunner(t *testing.T) *KubectlRunner {
	p, err := exec.LookPath(kubectlBinary)
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

func (k *KubectlRunner) RunCommand(args []string) (stdout []byte, err error) {
	inner := func() error {
		cmd := exec.Command(k.BinaryPath, args...)
		stdout, err = cmd.CombinedOutput()
		if err != nil {
			k.T.Logf("Error %s running command %s. Return code: %s", stdout, args, err)
			return &commonutil.RetriableError{Err: fmt.Errorf("Error running command. Error  %s. Output: %s", err, stdout)}
		}
		return nil
	}

	err = commonutil.RetryAfter(3, inner, 2*time.Second)
	return stdout, err
}

func (k *KubectlRunner) CreateRandomNamespace() string {
	const strLen = 20
	name := genRandString(strLen)
	if _, err := k.RunCommand([]string{"create", "namespace", name}); err != nil {
		k.T.Fatalf("Error creating namespace: %s", err)
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

func (k *KubectlRunner) DeleteNamespace(namespace string) error {
	_, err := k.RunCommand([]string{"delete", "namespace", namespace})
	return err
}

func WaitForBusyboxRunning(t *testing.T, namespace string) error {
	client, err := commonutil.GetClient()
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"integration-test": "busybox"}))
	return commonutil.WaitForPodsWithLabelRunning(client, namespace, selector)
}

func WaitForDashboardRunning(t *testing.T) error {
	client, err := commonutil.GetClient()
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}
	if err := commonutil.WaitForDeploymentToStabilize(client, "kube-system", "kubernetes-dashboard", time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for dashboard deployment to stabilize")
	}

	return nil
}

func WaitForIngressControllerRunning(t *testing.T) error {
	client, err := commonutil.GetClient()
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	if err := commonutil.WaitForDeploymentToStabilize(client, "kube-system", "nginx-ingress-controller", time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for ingress-controller deployment to stabilize")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"app": "nginx-ingress-controller"}))
	if err := commonutil.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "waiting for ingress-controller pods")
	}

	return nil
}

func WaitForIngressDefaultBackendRunning(t *testing.T) error {
	client, err := commonutil.GetClient()
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	if err := commonutil.WaitForDeploymentToStabilize(client, "kube-system", "default-http-backend", time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for default-http-backend deployment to stabilize")
	}

	if err := commonutil.WaitForService(client, "kube-system", "default-http-backend", true, time.Millisecond*500, time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for default-http-backend service to be up")
	}

	if err := commonutil.WaitForServiceEndpointsNum(client, "kube-system", "default-http-backend", 1, time.Second*3, time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for one default-http-backend endpoint to be up")
	}

	return nil
}

func WaitForNginxRunning(t *testing.T) error {
	client, err := commonutil.GetClient()

	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx"}))
	if err := commonutil.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		return errors.Wrap(err, "waiting for nginx pods")
	}

	if err := commonutil.WaitForService(client, "default", "nginx", true, time.Millisecond*500, time.Minute*10); err != nil {
		t.Errorf("Error waiting for nginx service to be up")
	}
	return nil
}

func Retry(t *testing.T, callback func() error, d time.Duration, attempts int) (err error) {
	for i := 0; i < attempts; i++ {
		err = callback()
		if err == nil {
			return nil
		}
		t.Logf("Error: %s, Retrying in %s. %d Retries remaining.", err, d, attempts-i)
		time.Sleep(d)
	}
	return err
}
