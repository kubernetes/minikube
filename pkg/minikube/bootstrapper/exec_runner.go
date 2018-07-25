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

package bootstrapper

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
)

// ExecRunner runs commands using the os/exec package.
//
// It implements the CommandRunner interface.
type ExecRunner struct{}

// Run starts the specified command in a bash shell and waits for it to complete.
func (*ExecRunner) Run(cmd string) error {
	glog.Infoln("Run:", cmd)
	c := exec.Command("/bin/bash", "-c", cmd)
	if err := c.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command `%s`", cmd)
	}
	return nil
}

// RunWithOutputTo runs the command (as in Run()) and redirects both its
// stdout and stderr to `out`.
func (*ExecRunner) RunWithOutputTo(cmd string, out io.Writer) error {
	glog.Infoln("Run with output:", cmd)
	c := exec.Command("/bin/bash", "-c", cmd)
	c.Stdout = out
	c.Stderr = out
	if err := c.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command `%s`", cmd)
	}

	return nil
}

// RunWithInputFrom() runs the command (as in Run()) and passes `in` as its stdin.
func (*ExecRunner) RunWithInputFrom(cmd string, in io.Reader) error {
	glog.Infoln("Run:", cmd)
	c := exec.Command("/bin/bash", "-c", cmd)
	c.Stdin = in
	if err := c.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command `%s`", cmd)
	}
	return nil
}

// RunWithOutput runs the command  in a bash shell and returns its
// combined standard output and standard error.
func (e *ExecRunner) RunWithOutput(cmd string) (string, error) {
	var b bytes.Buffer
	err := e.RunWithOutputTo(cmd, &b)
	if err != nil {
		return b.String(), err
	}
	return b.String(), nil
}

// Copy copies a file and its permissions
func (e *ExecRunner) Copy(f assets.CopyableFile) error {
	targetPath := filepath.Join(f.GetTargetDir(), f.GetTargetName())

	for _, cmd := range []string{getDeleteFileCommand(f), getMkDirCommand(f)} {
		if err := e.Run(cmd); err != nil {
			return errors.Wrapf(err, "failed to prepare to copy `%s`", targetPath)
		}
	}

	cmd := fmt.Sprintf("sudo install -Dm%s /dev/stdin '%s'", f.GetPermissions(), targetPath)
	if err := e.RunWithInputFrom(cmd, f); err != nil {
		return errors.Wrapf(err, "failed to write `%s`", targetPath)
	}
	return nil
}

// Remove removes a file
func (e *ExecRunner) Remove(f assets.CopyableFile) error {
	cmd := getDeleteFileCommand(f)
	if err := e.Run(cmd); err != nil {
		targetPath := filepath.Join(f.GetTargetDir(), f.GetTargetName())
		return errors.Wrapf(err, "failed to remove `%s`", targetPath)
	}
	return nil
}
