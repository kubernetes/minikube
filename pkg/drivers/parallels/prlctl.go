/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package parallels

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine/log"
)

func detectCmdInPath(cmd string) string {
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}

var (
	prlctlCmd      = detectCmdInPath("prlctl")
	prlsrvctlCmd   = detectCmdInPath("prlsrvctl")
	prldisktoolCmd = detectCmdInPath("prl_disk_tool")

	errPrlctlNotFound      = errors.New("Could not detect `prlctl` binary! Make sure Parallels Desktop Pro or Business edition is installed")
	errPrlsrvctlNotFound   = errors.New("Could not detect `prlsrvctl` binary! Make sure Parallels Desktop Pro or Business edition is installed")
	errPrldisktoolNotFound = errors.New("Could not detect `prl_disk_tool` binary! Make sure Parallels Desktop Pro or Business edition is installed")
)

func runCmd(cmdName string, args []string, notFound error) (string, string, error) {
	cmd := exec.Command(cmdName, args...)
	if os.Getenv("MACHINE_DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	log.Debugf("executing: %v %v", cmdName, strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee.Err == exec.ErrNotFound {
			err = notFound
		}
	}
	return stdout.String(), stderr.String(), err
}

func prlctl(args ...string) error {
	_, _, err := runCmd(prlctlCmd, args, errPrlctlNotFound)
	return err
}

func prlctlOutErr(args ...string) (string, string, error) {
	return runCmd(prlctlCmd, args, errPrlctlNotFound)
}

func prlsrvctl(args ...string) error {
	_, _, err := runCmd(prlsrvctlCmd, args, errPrlsrvctlNotFound)
	return err
}

func prlsrvctlOutErr(args ...string) (string, string, error) {
	return runCmd(prlsrvctlCmd, args, errPrlsrvctlNotFound)
}

func prldisktool(args ...string) error {
	_, _, err := runCmd(prldisktoolCmd, args, errPrldisktoolNotFound)
	return err
}
