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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
)

// ExecRunner runs commands using the os/exec package.
//
// It implements the CommandRunner interface.
type execRunner struct {
}

// NewExecRunner returns a kicRunner implementor of runner which runs cmds inside a container
func NewExecRunner() Runner {
	return &execRunner{}
}

// RunCmd implements the Command Runner interface to run a exec.Cmd object
func (*execRunner) RunCmd(cmd *exec.Cmd) (*RunResult, error) {
	rr := &RunResult{Args: cmd.Args}
	klog.Infof("Run: %v", rr.Command())

	var outb, errb io.Writer
	if cmd.Stdout == nil {
		var so bytes.Buffer
		outb = io.MultiWriter(&so, &rr.Stdout)
	} else {
		outb = io.MultiWriter(cmd.Stdout, &rr.Stdout)
	}

	if cmd.Stderr == nil {
		var se bytes.Buffer
		errb = io.MultiWriter(&se, &rr.Stderr)
	} else {
		errb = io.MultiWriter(cmd.Stderr, &rr.Stderr)
	}

	cmd.Stdout = outb
	cmd.Stderr = errb

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	if exitError, ok := err.(*exec.ExitError); ok {
		rr.ExitCode = exitError.ExitCode()
	}
	// Decrease log spam
	if elapsed > (1 * time.Second) {
		klog.Infof("Completed: %s: (%s)", rr.Command(), elapsed)
	}
	if err == nil {
		return rr, nil
	}

	return rr, fmt.Errorf("%s: %v\nstdout:\n%s\nstderr:\n%s", rr.Command(), err, rr.Stdout.String(), rr.Stderr.String())
}

// Copy copies a file and its permissions
func (*execRunner) Copy(f assets.CopyableFile) error {
	dst := path.Join(f.GetTargetDir(), f.GetTargetName())
	if _, err := os.Stat(dst); err == nil {
		klog.Infof("found %s, removing ...", dst)
		if err := os.Remove(dst); err != nil {
			return errors.Wrapf(err, "error removing file %s", dst)
		}
	}

	src := f.GetSourcePath()
	klog.Infof("cp: %s --> %s (%d bytes)", src, dst, f.GetLength())
	if f.GetLength() == 0 {
		klog.Warningf("0 byte asset: %+v", f)
	}

	perms, err := strconv.ParseInt(f.GetPermissions(), 8, 0)
	if err != nil {
		return errors.Wrapf(err, "error converting permissions %s to integer", f.GetPermissions())
	}

	return writeFile(dst, f, os.FileMode(perms))
}

// Remove removes a file
func (*execRunner) Remove(f assets.CopyableFile) error {
	dst := filepath.Join(f.GetTargetDir(), f.GetTargetName())
	klog.Infof("rm: %s", dst)
	return os.Remove(dst)
}
