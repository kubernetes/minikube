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

package sshutil

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/docker/docker/pkg/term"
	"github.com/docker/machine/libmachine/drivers"
	machinessh "github.com/docker/machine/libmachine/ssh"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/util"
)

// SSHSession provides methods for running commands on a host.
type SSHSession interface {
	Close() error
	StdinPipe() (io.WriteCloser, error)
	Run(cmd string) error
	Wait() error
}

// NewSSHClient returns an SSH client object for running commands.
func NewSSHClient(d drivers.Driver) (*ssh.Client, error) {
	h, err := newSSHHost(d)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new ssh host from driver")

	}
	auth := &machinessh.Auth{}
	if h.SSHKeyPath != "" {
		auth.Keys = []string{h.SSHKeyPath}
	}
	config, err := machinessh.NewNativeConfig(h.Username, auth)
	if err != nil {
		return nil, errors.Wrapf(err, "Error creating new native config from ssh using: %s, %s", h.Username, auth)
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(h.IP, strconv.Itoa(h.Port)), &config)
	if err != nil {
		return nil, errors.Wrap(err, "Error dialing tcp via ssh client")
	}
	return client, nil
}

func DeleteAddon(a *assets.Addon, client *ssh.Client) error {
	m := util.MultiError{}
	for _, f := range a.Assets {
		if err := DeleteFile(f, client); err != nil {
			m.Collect(err)
		}
	}
	return m.ToError()
}

func TransferAddon(a *assets.Addon, client *ssh.Client) error {
	m := util.MultiError{}
	for _, f := range a.Assets {
		if err := TransferFile(f, client); err != nil {
			m.Collect(err)
		}
	}
	return m.ToError()
}

func TransferFile(f assets.CopyableFile, client *ssh.Client) error {
	return Transfer(f, f.GetLength(),
		f.GetTargetDir(), f.GetTargetName(),
		f.GetPermissions(), client)
}

// Transfer uses an SSH session to copy a file to the remote machine.
func Transfer(reader io.Reader, readerLen int, remotedir, filename string, perm string, c *ssh.Client) error {
	// Delete the old file first. This makes sure permissions get reset.
	deleteCmd := fmt.Sprintf("sudo rm -f %s", filepath.Join(remotedir, filename))
	mkdirCmd := fmt.Sprintf("sudo mkdir -p %s", remotedir)
	for _, cmd := range []string{deleteCmd, mkdirCmd} {
		glog.Infoln("Running: ", cmd)
		if err := RunCommand(c, cmd); err != nil {
			return errors.Wrapf(err, "Error running command: %s", cmd)
		}
	}

	s, err := c.NewSession()
	if err != nil {
		return errors.Wrap(err, "Error creating new session via ssh client")
	}

	w, err := s.StdinPipe()
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
		header := fmt.Sprintf("C%s %d %s\n", perm, readerLen, filename)
		fmt.Fprint(w, header)
		io.Copy(w, reader)
		fmt.Fprint(w, "\x00")
	}()

	scpcmd := fmt.Sprintf("sudo scp -t %s", remotedir)
	if err := s.Run(scpcmd); err != nil {
		return errors.Wrapf(err, "Error running scp command: %s", scpcmd)
	}
	wg.Wait()

	return nil
}

func RunCommand(c *ssh.Client, cmd string) error {
	s, err := c.NewSession()
	defer s.Close()
	glog.Infoln("Running command: %s", cmd)
	if err != nil {
		return errors.Wrap(err, "Error creating new session for ssh client")
	}

	return s.Run(cmd)
}

func RunCommandOutput(c *ssh.Client, cmd string) (string, error) {
	s, err := c.NewSession()
	defer s.Close()
	glog.Infoln("Running command: %s", cmd)
	if err != nil {
		return "", errors.Wrap(err, "Error creating new session for ssh client")
	}
	b, err := s.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrap(err, "Running ssh command")
	}
	glog.Infoln("Command output: %s", string(b))
	return string(b), nil
}

type sshHost struct {
	IP         string
	Port       int
	SSHKeyPath string
	Username   string
}

func newSSHHost(d drivers.Driver) (*sshHost, error) {

	ip, err := d.GetSSHHostname()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ssh host name for driver")
	}
	port, err := d.GetSSHPort()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ssh port for driver")
	}
	return &sshHost{
		IP:         ip,
		Port:       port,
		SSHKeyPath: d.GetSSHKeyPath(),
		Username:   d.GetSSHUsername(),
	}, nil
}

func DeleteFile(f assets.CopyableFile, client *ssh.Client) error {
	return RunCommand(client, GetDeleteFileCommand(f))
}

func GetDeleteFileCommand(f assets.CopyableFile) string {
	return fmt.Sprintf("sudo rm %s", filepath.Join(f.GetTargetDir(), f.GetTargetName()))
}

func GetShell(session *ssh.Session, cmd string) error {
	var (
		termWidth, termHeight int
	)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}
	fd := os.Stdin.Fd()
	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}

		defer term.RestoreTerminal(fd, oldState)

		winsize, err := term.GetWinsize(fd)
		if err != nil {
			termWidth = 80
			termHeight = 24
		} else {
			termWidth = int(winsize.Width)
			termHeight = int(winsize.Height)
		}
	}
	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return err
	}
	session.Run(cmd)

	return nil
}
