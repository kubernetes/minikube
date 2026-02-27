/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package node

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"

	"k8s.io/klog/v2"
)

var powershell string

var (
	ErrPowerShellNotFound = errors.New("powershell was not found in the path")
	ErrNotAdministrator   = errors.New("hyper-v commands have to be run as an Administrator")
	ErrNotInstalled       = errors.New("hyper-V PowerShell Module is not available")
)

func init() {
	powershell, _ = exec.LookPath("powershell.exe")
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
	return stdout.String(), err
}

func cmd(args ...string) error {
	_, err := cmdOut(args...)
	return err
}

func CmdOutSSH(client *ssh.Client, script string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	command := fmt.Sprintf("powershell -NoProfile -NonInteractive -Command \"%s\"", script)
	klog.Infof("[executing] : %v", command)

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	klog.Infof("[stdout =====>] : %s", stdout.String())
	klog.Infof("[stderr =====>] : %s", stderr.String())
	return stdout.String(), err
}

func cmdSSH(client *ssh.Client, args ...string) error {
	script := strings.Join(args, " ")
	_, err := CmdOutSSH(client, script)
	return err
}
