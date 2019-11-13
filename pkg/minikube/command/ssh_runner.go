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
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/util"
)

var (
	layout = "2006-01-02 15:04:05.999999999 -0700"
)

// SSHRunner runs commands through SSH.
//
// It implements the CommandRunner interface.
type SSHRunner struct {
	c *ssh.Client
}

// NewSSHRunner returns a new SSHRunner that will run commands
// through the ssh.Client provided.
func NewSSHRunner(c *ssh.Client) *SSHRunner {
	return &SSHRunner{c}
}

// Remove runs a command to delete a file on the remote.
func (s *SSHRunner) Remove(f assets.CopyableFile) error {
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	cmd := getDeleteFileCommand(f)
	return sess.Run(cmd)
}

// teeSSH runs an SSH command, streaming stdout, stderr to logs
func teeSSH(s *ssh.Session, cmd string, outB io.Writer, errB io.Writer) error {
	outPipe, err := s.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "stdout")
	}

	errPipe, err := s.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "stderr")
	}
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if err := teePrefix(util.ErrPrefix, errPipe, errB, glog.V(8).Infof); err != nil {
			glog.Errorf("tee stderr: %v", err)
		}
		wg.Done()
	}()
	go func() {
		if err := teePrefix(util.OutPrefix, outPipe, outB, glog.V(8).Infof); err != nil {
			glog.Errorf("tee stdout: %v", err)
		}
		wg.Done()
	}()
	err = s.Run(cmd)
	wg.Wait()
	return err
}

// RunCmd implements the Command Runner interface to run a exec.Cmd object
func (s *SSHRunner) RunCmd(cmd *exec.Cmd) (*RunResult, error) {
	rr := &RunResult{Args: cmd.Args}
	glog.Infof("(SSHRunner) Run:  %v", rr.Command())

	var outb, errb io.Writer
	start := time.Now()

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

	sess, err := s.c.NewSession()
	if err != nil {
		return rr, errors.Wrap(err, "NewSession")
	}

	defer func() {
		if err := sess.Close(); err != nil {
			if err != io.EOF {
				glog.Errorf("session close: %v", err)
			}
		}
	}()

	elapsed := time.Since(start)
	err = teeSSH(sess, shellquote.Join(cmd.Args...), outb, errb)
	if err == nil {
		// Reduce log spam
		if elapsed > (1 * time.Second) {
			glog.Infof("(SSHRunner) Done: %v: (%s)", rr.Command(), elapsed)
		}
	} else {
		if exitError, ok := err.(*exec.ExitError); ok {
			rr.ExitCode = exitError.ExitCode()
		}
		glog.Infof("(SSHRunner) Non-zero exit: %v: %v (%s)\n%s", rr.Command(), err, elapsed, rr.Output())
	}
	return rr, err
}

// Copy copies a file to the remote over SSH.
func (s *SSHRunner) Copy(f assets.CopyableFile) error {
	dst := path.Join(path.Join(f.GetTargetDir(), f.GetTargetName()))
	exists, err := s.sameFileExists(f, dst)
	if err != nil {
		glog.Infof("Checked if %s exists, but got error: %v", dst, err)
	}
	if exists {
		glog.Infof("Skipping copying %s as it already exists", dst)
		return nil
	}
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "NewSession")
	}

	w, err := sess.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "StdinPipe")
	}
	// The scpcmd below *should not* return until all data is copied and the
	// StdinPipe is closed. But let's use errgroup to make it explicit.
	var g errgroup.Group
	var copied int64
	glog.Infof("Transferring %d bytes to %s", f.GetLength(), dst)

	g.Go(func() error {
		defer w.Close()
		header := fmt.Sprintf("C%s %d %s\n", f.GetPermissions(), f.GetLength(), f.GetTargetName())
		fmt.Fprint(w, header)
		if f.GetLength() == 0 {
			glog.Warningf("%s is a 0 byte asset!", f.GetTargetName())
			fmt.Fprint(w, "\x00")
			return nil
		}

		copied, err = io.Copy(w, f)
		if err != nil {
			return errors.Wrap(err, "io.Copy")
		}
		if copied != int64(f.GetLength()) {
			return fmt.Errorf("%s: expected to copy %d bytes, but copied %d instead", f.GetTargetName(), f.GetLength(), copied)
		}
		glog.Infof("%s: copied %d bytes", f.GetTargetName(), copied)
		fmt.Fprint(w, "\x00")
		return nil
	})

	scp := fmt.Sprintf("sudo mkdir -p %s && sudo scp -t %s", f.GetTargetDir(), f.GetTargetDir())
	mtime, err := f.GetModTime()
	if err != nil {
		glog.Infof("error getting modtime for %s: %v", dst, err)
	} else {
		scp += fmt.Sprintf(" && sudo touch -d \"%s\" %s", mtime.Format(layout), dst)
	}
	out, err := sess.CombinedOutput(scp)
	if err != nil {
		return fmt.Errorf("%s: %s\noutput: %s", scp, err, out)
	}
	return g.Wait()
}

func (s *SSHRunner) sameFileExists(f assets.CopyableFile, dst string) (bool, error) {
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
	sess, err := s.c.NewSession()
	if err != nil {
		return false, err
	}

	cmd := "stat -c \"%s %y\" " + dst
	out, err := sess.CombinedOutput(cmd)
	if err != nil {
		return false, err
	}
	outputs := strings.SplitN(strings.Trim(string(out), "\n"), " ", 2)

	dstSize, err := strconv.Atoi(outputs[0])
	if err != nil {
		return false, err
	}
	dstModTime, err := time.Parse(layout, outputs[1])
	if err != nil {
		return false, err
	}

	// compare sizes and modtimes
	if srcSize != dstSize {
		return false, errors.New("source file and destination file are different sizes")
	}
	return srcModTime.Equal(dstModTime), nil
}

// teePrefix copies bytes from a reader to writer, logging each new line.
func teePrefix(prefix string, r io.Reader, w io.Writer, logger func(format string, args ...interface{})) error {
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
