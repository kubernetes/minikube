// +build windows

/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package hyperv

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

var powershell string

func init() {
	powershell, _ = exec.LookPath("powershell")
}

func cmdOut(args ...string) (string, error) {
	args = append([]string{"-NoProfile", "-NonInteractive"}, args...)
	cmd := exec.Command(powershell, args...)
	klog.Infof("[executing ==>] : %v %v", powershell, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	klog.Infof("[stdout =====>] : %s", stdout.String())
	klog.Infof("[stderr =====>] : %s", stderr.String())
	if err != nil {
		klog.Infof("[err =====>] : %v", err)
	}
	return stdout.String(), err
}

func cmd(args ...string) error {
	_, err := cmdOut(args...)
	return err
}

func parseLines(stdout string) []string {
	var resp []string

	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		resp = append(resp, s.Text())
	}

	return resp
}
