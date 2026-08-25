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

package tests

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"maps"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// CommandResult defines the mock response for a command.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// SSHServer is a test SSH server returning preconfigured responses.
// Runs one session at a time, matching minikube's SSHRunner.
type SSHServer struct {
	Config    *ssh.ServerConfig
	ClientKey ssh.Signer

	commands map[string]CommandResult
	quit     atomic.Bool
	port     int
	listener net.Listener
	t        *testing.T
}

// NewSSHServer returns a new SSHServer instance with ed25519 key auth
// and the given command responses.
func NewSSHServer(t *testing.T, commands map[string]CommandResult) (*SSHServer, error) {
	t.Helper()

	_, clientKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client key: %w", err)
	}
	clientSigner, err := ssh.NewSignerFromKey(clientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create client signer: %w", err)
	}
	clientPub := clientSigner.PublicKey()

	_, hostKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate host key: %w", err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create host signer: %w", err)
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Equal(key.Marshal(), clientPub.Marshal()) {
				return nil, nil
			}
			return nil, errors.New("unknown key")
		},
	}
	config.AddHostKey(hostSigner)

	return &SSHServer{
		Config:    config,
		ClientKey: clientSigner,
		commands:  maps.Clone(commands),
		t:         t,
	}, nil
}

// Start the SSH server on a random port.
func (s *SSHServer) Start() error {
	if s.port != 0 {
		return errors.New("server already started")
	}
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = l
	s.port = l.Addr().(*net.TCPAddr).Port
	s.t.Logf("Listening on 127.0.0.1:%d", s.port)
	go s.serve()
	return nil
}

// Dial connects to the server using the client key. Returns an ssh.Client.
func (s *SSHServer) Dial() (*ssh.Client, error) {
	if s.port == 0 {
		return nil, errors.New("server not started")
	}
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	return ssh.Dial("tcp", addr, &ssh.ClientConfig{
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(s.ClientKey)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
}

// Stop the SSH server. Safe to call multiple times.
func (s *SSHServer) Stop() {
	if s.quit.Swap(true) {
		return
	}
	s.t.Logf("Stopping")
	s.listener.Close()
}

// Private implementation below.

type execRequest struct {
	Command string
}

func (s *SSHServer) serve() {
	for {
		c, err := s.listener.Accept()
		if s.quit.Load() {
			return
		}
		if err != nil {
			s.t.Errorf("Failed to accept: %v", err)
			return
		}
		s.handleConnection(c)
	}
}

func (s *SSHServer) handleConnection(c net.Conn) {
	_, chans, reqs, err := ssh.NewServerConn(c, s.Config)
	if err != nil {
		s.t.Logf("NewServerConn: %v", err)
		return
	}
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		channel, requests, err := newChannel.Accept()
		if err != nil {
			s.t.Logf("channel accept: %v", err)
			return
		}
		for req := range requests {
			s.handleRequest(channel, req)
		}
	}
}

func (s *SSHServer) handleRequest(channel ssh.Channel, req *ssh.Request) {
	switch req.Type {
	case "exec":
		s.doExec(channel, req)
	case "pty-req":
		s.doPtyReq(req)
	default:
		s.doUnknown(channel, req)
	}
}

func (s *SSHServer) doExec(channel ssh.Channel, req *ssh.Request) {
	defer channel.Close()

	if err := req.Reply(true, nil); err != nil {
		s.t.Errorf("Failed to reply: %v", err)
		return
	}

	var cmd execRequest
	if err := ssh.Unmarshal(req.Payload, &cmd); err != nil {
		s.t.Errorf("Failed to unmarshal: %v", err)
		return
	}

	result, ok := s.commands[cmd.Command]
	if !ok {
		result = CommandResult{
			Stderr:   fmt.Sprintf("%s: command not found", cmd.Command),
			ExitCode: 127,
		}
	}

	if result.Stdout != "" {
		if _, err := channel.Write([]byte(result.Stdout)); err != nil {
			s.t.Errorf("Failed to write stdout: %v", err)
			return
		}
	}
	if result.Stderr != "" {
		if _, err := channel.Stderr().Write([]byte(result.Stderr)); err != nil {
			s.t.Errorf("Failed to write stderr: %v", err)
			return
		}
	}

	exitBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(exitBytes, uint32(result.ExitCode))
	if _, err := channel.SendRequest("exit-status", false, exitBytes); err != nil {
		s.t.Errorf("Failed to send exit-status: %v", err)
	}
}

func (s *SSHServer) doPtyReq(req *ssh.Request) {
	if err := req.Reply(true, nil); err != nil {
		s.t.Errorf("Failed to reply: %v", err)
	}
}

func (s *SSHServer) doUnknown(channel ssh.Channel, req *ssh.Request) {
	s.t.Errorf("Unexpected request type: %s", req.Type)
	if req.WantReply {
		req.Reply(false, nil)
	}
	channel.Close()
}
