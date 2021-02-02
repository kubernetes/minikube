/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/tests"
)

func TestRunCmdWarnSlowOnce(t *testing.T) {
	if runtime.GOOS != "linux" {
		return
	}
	f1 := tests.NewFakeFile()
	out.SetErrFile(f1)

	cmd := exec.Command("sleep", "3")
	_, err := runCmd(cmd, true)

	if err != nil {
		t.Errorf("runCmd has error: %v", err)
	}

	if !strings.Contains(f1.String(), "Executing \"sleep 3\" took an unusually long time") {
		t.Errorf("runCmd does not print the correct log, instead print :%v", f1.String())
	}

	f2 := tests.NewFakeFile()
	out.SetErrFile(f2)

	cmd = exec.Command("sleep", "3")
	_, err = runCmd(cmd, true)

	if err != nil {
		t.Errorf("runCmd has error: %v", err)
	}

	if strings.Contains(f2.String(), "Executing \"sleep 3\" took an unusually long time") {
		t.Errorf("runCmd does not print the correct log, instead print :%v", f2.String())
	}
}
