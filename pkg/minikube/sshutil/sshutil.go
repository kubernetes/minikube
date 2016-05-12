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
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/docker/machine/libmachine/drivers"
	machinessh "github.com/docker/machine/libmachine/ssh"
	"golang.org/x/crypto/ssh"
)

// SSHSession provides methods for running commands on a host.
type SSHSession interface {
	Close() error
	StdinPipe() (io.WriteCloser, error)
	Start(cmd string) error
	Wait() error
}

// NewSSHSession returns an SSHSession object for running commands.
func NewSSHSession(d drivers.Driver) (SSHSession, error) {
	h, err := newSSHHost(d)
	if err != nil {
		return nil, err

	}
	auth := &machinessh.Auth{}
	if h.SSHKeyPath != "" {
		auth.Keys = []string{h.SSHKeyPath}
	}
	config, err := machinessh.NewNativeConfig(h.Username, auth)
	if err != nil {
		return nil, err
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", h.IP, h.Port), &config)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Transfer uses an SSH session to copy a file to the remote machine.
func Transfer(localpath, remotepath string, r SSHSession) error {
	f, err := os.Open(localpath)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)

	cmd := fmt.Sprintf("cat > %s", remotepath)
	stdin, err := r.StdinPipe()
	if err != nil {
		return err
	}

	if err := r.Start(cmd); err != nil {
		return err
	}
	_, err = io.Copy(stdin, reader)
	stdin.Close()
	if err != nil {
		return err
	}

	return r.Wait()
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
		return nil, err
	}
	port, err := d.GetSSHPort()
	if err != nil {
		return nil, err
	}
	return &sshHost{
		IP:         ip,
		Port:       port,
		SSHKeyPath: d.GetSSHKeyPath(),
		Username:   d.GetSSHUsername(),
	}, nil
}
