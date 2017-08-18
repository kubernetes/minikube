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

type SSHRunner struct {
	c *ssh.Client
}

func NewSSHRunner(c *ssh.Client) *SSHRunner {
	return &SSHRunner{c}
}

func (s *SSHRunner) Remove(f assets.CopyableFile) error {
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	cmd := getDeleteFileCommand(f)
	return sess.Run(cmd)
}

func (s *SSHRunner) Run(cmd string) error {
	glog.Infoln("Run:", cmd)
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	return sess.Run(cmd)
}

func (s *SSHRunner) CombinedOutput(cmd string) (string, error) {
	glog.Infoln("Run with output:", cmd)
	sess, err := s.c.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	out, err := sess.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "running command: %s\n output: %s", cmd, out)
	}
	return string(out), nil
}

func (s *SSHRunner) Copy(f assets.CopyableFile) error {
	deleteCmd := fmt.Sprintf("sudo rm -f %s", filepath.Join(f.GetTargetDir(), f.GetTargetName()))
	mkdirCmd := fmt.Sprintf("sudo mkdir -p %s", f.GetTargetDir())
	for _, cmd := range []string{deleteCmd, mkdirCmd} {
		if err := s.Run(cmd); err != nil {
			return errors.Wrapf(err, "Error running command: %s", cmd)
		}
	}

	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "Error creating new session via ssh client")
	}

	w, err := sess.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "Error accessing StdinPipe via ssh session")
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
		return errors.Wrapf(err, "Error running scp command: %s", scpcmd)
	}
	wg.Wait()

	return nil
}
