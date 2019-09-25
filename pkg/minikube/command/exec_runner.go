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

package command

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	osexec "os/exec"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
)

type ExecCmd struct {
	Cmd *osexec.Cmd
}

func (ec *ExecCmd) Run() error {
	glog.Infoln("Run:", ec.Cmd.Args)
	err := ec.Cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "running command: %s", ec.Cmd.Args)
	}
	return err
}

func (ec *ExecCmd) SetEnv(envs ...string) Cmd {
	ec.Cmd.Env = envs
	return ec
}

func (ec *ExecCmd) SetStdin(r io.Reader) Cmd {
	ec.Cmd.Stdin = r
	return ec
}

func (ec *ExecCmd) SetStdout(w io.Writer) Cmd {
	ec.Cmd.Stdout = w
	return ec
}

// SetStderr sets stderr
func (ec *ExecCmd) SetStderr(w io.Writer) Cmd {
	ec.Cmd.Stderr = w
	return ec
}

// ExecRunner runs commands using the os/exec package.
//
// It implements the CommandRunner interface.
type ExecRunner struct{}

// Run starts the specified command in a bash shell and waits for it to complete.
func (*ExecRunner) Command(cmd string) Cmd {
	c := exec.Command("/bin/bash", "-c", cmd)
	return &ExecCmd{
		Cmd: c,
	}
}

// Run starts the specified command in a bash shell and waits for it to complete.
func (*ExecRunner) Run(c Cmd) error {
	return c.Run()
}

// CombinedOutputTo runs the command and stores both command
// output and error to out.
func (*ExecRunner) CombinedOutputTo(c Cmd, out io.Writer) error {
	c.SetStdout(out)
	c.SetStderr(out)
	return c.Run()
}

// CombinedOutput runs the command  in a bash shell and returns its
// combined standard output and standard error.
func (e *ExecRunner) CombinedOutput(c Cmd) (string, error) {
	var b bytes.Buffer
	err := e.CombinedOutputTo(c, &b)
	return b.String(), err

}

// Copy copies a file and its permissions
func (*ExecRunner) Copy(f assets.CopyableFile) error {
	if err := os.MkdirAll(f.GetTargetDir(), os.ModePerm); err != nil {
		return errors.Wrapf(err, "error making dirs for %s", f.GetTargetDir())
	}
	targetPath := path.Join(f.GetTargetDir(), f.GetTargetName())
	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Remove(targetPath); err != nil {
			return errors.Wrapf(err, "error removing file %s", targetPath)
		}

	}
	target, err := os.Create(targetPath)
	if err != nil {
		return errors.Wrapf(err, "error creating file at %s", targetPath)
	}
	perms, err := strconv.ParseInt(f.GetPermissions(), 8, 0)
	if err != nil {
		return errors.Wrapf(err, "error converting permissions %s to integer", f.GetPermissions())
	}
	if err := os.Chmod(targetPath, os.FileMode(perms)); err != nil {
		return errors.Wrapf(err, "error changing file permissions for %s", targetPath)
	}

	if _, err = io.Copy(target, f); err != nil {
		return errors.Wrapf(err, `error copying file %s to target location:
do you have the correct permissions?`,
			targetPath)
	}
	return target.Close()
}

// Remove removes a file
func (e *ExecRunner) Remove(f assets.CopyableFile) error {
	targetPath := filepath.Join(f.GetTargetDir(), f.GetTargetName())
	return os.Remove(targetPath)
}
