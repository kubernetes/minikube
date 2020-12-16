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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
)

var (
	// ErrPrefix notes an error
	ErrPrefix = "! "

	// OutPrefix notes output
	OutPrefix = "> "

	// Mutex protects teePrefix from writing to same log buffer parallelly
	logMutex = &sync.Mutex{}
)

// RunResult holds the results of a Runner
type RunResult struct {
	Stdout   bytes.Buffer
	Stderr   bytes.Buffer
	ExitCode int
	Args     []string // the args that was passed to Runner
}

// StartedCmd holds the contents of a started command
type StartedCmd struct {
	cmd *exec.Cmd
	rr  *RunResult
}

// Runner represents an interface to run commands.
type Runner interface {
	// RunCmd runs a cmd of exec.Cmd type. allowing user to set cmd.Stdin, cmd.Stdout,...
	// not all implementors are guaranteed to handle all the properties of cmd.
	RunCmd(cmd *exec.Cmd) (*RunResult, error)

	// StartCmd starts a cmd of exec.Cmd type.
	// This func in non-blocking, use WaitCmd to block until complete.
	// Not all implementors are guaranteed to handle all the properties of cmd.
	StartCmd(cmd *exec.Cmd) (*StartedCmd, error)

	// WaitCmd will prevent further execution until the started command has completed.
	WaitCmd(startedCmd *StartedCmd) (*RunResult, error)

	// Copy is a convenience method that runs a command to copy a file
	Copy(assets.CopyableFile) error

	// Remove is a convenience method that runs a command to remove a file
	Remove(assets.CopyableFile) error
}

// Command returns a human readable command string that does not induce eye fatigue
func (rr RunResult) Command() string {
	var sb strings.Builder
	sb.WriteString(rr.Args[0])
	for _, a := range rr.Args[1:] {
		if strings.Contains(a, " ") {
			sb.WriteString(fmt.Sprintf(` "%s"`, a))
			continue
		}
		sb.WriteString(fmt.Sprintf(" %s", a))
	}
	return sb.String()
}

// Output returns human-readable output for an execution result
func (rr RunResult) Output() string {
	var sb strings.Builder
	if rr.Stdout.Len() > 0 {
		sb.WriteString(fmt.Sprintf("-- stdout --\n%s\n-- /stdout --", rr.Stdout.Bytes()))
	}
	if rr.Stderr.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n** stderr ** \n%s\n** /stderr **", rr.Stderr.Bytes()))
	}
	return sb.String()
}

// teePrefix copies bytes from a reader to writer, logging each new line.
func teePrefix(prefix string, r io.Reader, w io.Writer, logger func(format string, args ...interface{})) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)
	var line bytes.Buffer

	for scanner.Scan() {
		b := scanner.Bytes()
		if _, err := w.Write(b); err != nil {
			return err
		}
		if bytes.IndexAny(b, "\r\n") == 0 {
			if line.Len() > 0 {
				logger("%s%s", prefix, line.String())
				line.Reset()
			}
			continue
		}
		line.Write(b)
	}
	// Catch trailing output in case stream does not end with a newline
	if line.Len() > 0 {
		logger("%s%s", prefix, line.String())
	}
	return nil
}

// fileExists checks that the same file exists on the other end
func fileExists(r Runner, f assets.CopyableFile, dst string) (bool, error) {
	// It's too difficult to tell if the file exists with the exact contents
	if f.GetSourcePath() == assets.MemorySource {
		return false, nil
	}

	// get file size and modtime of the source
	srcSize := f.GetLength()
	srcModTime, err := f.GetModTime()
	if err != nil {
		return false, err
	}
	if srcModTime.IsZero() {
		return false, nil
	}

	// get file size and modtime of the destination
	rr, err := r.RunCmd(exec.Command("stat", "-c", "%s %y", dst))
	if err != nil {
		if rr.ExitCode == 1 {
			return false, nil
		}

		// avoid the noise because ssh doesn't propagate the exit code
		if strings.HasSuffix(err.Error(), "status 1") {
			return false, nil
		}

		return false, err
	}

	stdout := strings.TrimSpace(rr.Stdout.String())
	outputs := strings.SplitN(stdout, " ", 2)
	dstSize, err := strconv.Atoi(outputs[0])
	if err != nil {
		return false, err
	}

	dstModTime, err := time.Parse(layout, outputs[1])
	if err != nil {
		return false, err
	}

	if srcSize != dstSize {
		return false, errors.New("source file and destination file are different sizes")
	}

	return srcModTime.Equal(dstModTime), nil
}

// writeFile is like ioutil.WriteFile, but does not require reading file into memory
func writeFile(dst string, f assets.CopyableFile, perms os.FileMode) error {
	w, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, perms)
	if err != nil {
		return errors.Wrap(err, "create")
	}
	defer w.Close()

	r := f.(io.Reader)
	n, err := io.Copy(w, r)
	if err != nil {
		return errors.Wrap(err, "copy")
	}

	if n != int64(f.GetLength()) {
		return fmt.Errorf("%s: expected to write %d bytes, but wrote %d instead", dst, f.GetLength(), n)
	}
	return w.Close()
}
