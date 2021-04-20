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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
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
	sudo bool
}

// NewExecRunner returns a kicRunner implementor of runner which runs cmds inside a container
func NewExecRunner(sudo bool) Runner {
	return &execRunner{sudo: sudo}
}

// RunCmd implements the Command Runner interface to run a exec.Cmd object
func (e *execRunner) RunCmd(cmd *exec.Cmd) (*RunResult, error) {
	rr := &RunResult{Args: cmd.Args}
	klog.Infof("Run: %v", rr.Command())

	if e.sudo && runtime.GOOS != "linux" {
		return nil, fmt.Errorf("sudo not supported on %s", runtime.GOOS)
	}

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

// StartCmd implements the Command Runner interface to start a exec.Cmd object
func (*execRunner) StartCmd(cmd *exec.Cmd) (*StartedCmd, error) {
	rr := &RunResult{Args: cmd.Args}
	sc := &StartedCmd{cmd: cmd, rr: rr}
	klog.Infof("Start: %v", rr.Command())

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

	if err := cmd.Start(); err != nil {
		return sc, errors.Wrap(err, "start")
	}

	return sc, nil
}

// WaitCmd implements the Command Runner interface to wait until a started exec.Cmd object finishes
func (*execRunner) WaitCmd(sc *StartedCmd) (*RunResult, error) {
	rr := sc.rr

	err := sc.cmd.Wait()
	if exitError, ok := err.(*exec.ExitError); ok {
		rr.ExitCode = exitError.ExitCode()
	}

	if err == nil {
		return rr, nil
	}

	return rr, fmt.Errorf("%s: %v\nstdout:\n%s\nstderr:\n%s", rr.Command(), err, rr.Stdout.String(), rr.Stderr.String())
}

// Copy copies a file and its permissions
func (e *execRunner) Copy(f assets.CopyableFile) error {
	dst := path.Join(f.GetTargetDir(), f.GetTargetName())
	if _, err := os.Stat(dst); err == nil {
		klog.Infof("found %s, removing ...", dst)
		if err := e.Remove(f); err != nil {
			return errors.Wrapf(err, "error removing file %s", dst)
		}
	}

	src := f.GetSourcePath()
	klog.Infof("cp: %s --> %s (%d bytes)", src, dst, f.GetLength())
	if f.GetLength() == 0 {
		klog.Warningf("0 byte asset: %+v", f)
	}

	perms, err := strconv.ParseInt(f.GetPermissions(), 8, 0)
	if err != nil || perms > 07777 {
		return errors.Wrapf(err, "error converting permissions %s to integer", f.GetPermissions())
	}

	if e.sudo {
		// write to temp location ...
		tmpfile, err := ioutil.TempFile("", "minikube")
		if err != nil {
			return errors.Wrapf(err, "error creating tempfile")
		}
		defer os.Remove(tmpfile.Name())
		err = writeFile(tmpfile.Name(), f, os.FileMode(perms))
		if err != nil {
			return errors.Wrapf(err, "error writing to tempfile %s", tmpfile.Name())
		}

		// ... then use sudo to move to target ...
		_, err = e.RunCmd(exec.Command("sudo", "cp", "-a", tmpfile.Name(), dst))
		if err != nil {
			return errors.Wrapf(err, "error copying tempfile %s to dst %s", tmpfile.Name(), dst)
		}

		// ... then fix file permission that should have been fine because of "cp -a"
		err = os.Chmod(dst, os.FileMode(perms))
		return err
	}
	return writeFile(dst, f, os.FileMode(perms))
}

// Remove removes a file
func (e *execRunner) Remove(f assets.CopyableFile) error {
	dst := filepath.Join(f.GetTargetDir(), f.GetTargetName())
	klog.Infof("rm: %s", dst)
	if e.sudo {
		if err := os.Remove(dst); err != nil {
			if !os.IsPermission(err) {
				return err
			}
			_, err = e.RunCmd(exec.Command("sudo", "rm", "-f", dst))
			if err != nil {
				return err
			}
		}
		return nil
	}
	return os.Remove(dst)
}
