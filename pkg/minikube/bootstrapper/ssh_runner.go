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
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"k8s.io/minikube/pkg/minikube/assets"
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

// Run starts a command on the remote in a shell and waits for it to return.
func (s *SSHRunner) Run(cmd string) error {
	glog.Infoln("Run:", cmd)
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create a SSH session")
	}
	defer sess.Close()
	return sess.Run(cmd)
}

// RunWithOutputTo runs the command (as in Run()) and writes both its stdout
// and stderr to `out`.
func (s *SSHRunner) RunWithOutputTo(cmd string, out io.Writer) error {
	b, err := s.RunWithOutput(cmd)
	if err != nil {
		return errors.Wrapf(err, "failed to run command `%s`", cmd)
	}
	_, err = out.Write([]byte(b))
	return err
}

// RunWithOutput runs the command (as in Run()) and returns its combined
// standard output and standard error.
func (s *SSHRunner) RunWithOutput(cmd string) (string, error) {
	glog.Infoln("Run with output:", cmd)
	sess, err := s.c.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "failed to create a SSH session")
	}
	defer sess.Close()

	b, err := sess.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "failed to run command `%s`", cmd)
	}
	return string(b), nil
}

// Copy copies a file to the remote over SSH.
func (s *SSHRunner) Copy(f assets.CopyableFile) error {
	targetPath := filepath.Join(f.GetTargetDir(), f.GetTargetName())

	for _, cmd := range []string{getDeleteFileCommand(f), getMkDirCommand(f)} {
		if err := s.Run(cmd); err != nil {
			return errors.Wrapf(err, "failed to prepare to copy file `%s`", targetPath)
		}
	}

	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create a SSH session")
	}

	w, err := sess.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "failed to access StdinPipe of a SSH session")
	}
	// The scpcmd below *should not* return until all data is copied and the
	// StdinPipe is closed. But let's use a WaitGroup to make it expicit.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer w.Close()
		header := fmt.Sprintf("C%s %d %s\n", f.GetPermissions(), f.GetLength(), f.GetTargetName())
		fmt.Fprint(w, header)
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()

	scpcmd := fmt.Sprintf("sudo scp -t %s", f.GetTargetDir())
	if err := sess.Run(scpcmd); err != nil {
		return errors.Wrapf(err, "failed to copy file `%s` via scp", targetPath)
	}
	wg.Wait()

	return nil
}
