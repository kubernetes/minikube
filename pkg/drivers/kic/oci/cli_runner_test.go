package oci

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/tests"
)

func TestCliRunnerOnlyPrintOnce(t *testing.T) {
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
