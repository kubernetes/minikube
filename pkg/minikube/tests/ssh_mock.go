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
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// SSHServer provides a mock SSH Server for testing. Commands are stored, not executed.
type SSHServer struct {
	Config *ssh.ServerConfig
	// Commands stores the raw commands executed against the server.
	Commands  map[string]int
	Connected bool
	Transfers *bytes.Buffer
	// Only access this with atomic ops
	hadASessionRequested int32
	// commandToOutput can be used to mock what the SSHServer returns for a given command
	// Only access this with atomic ops
	commandToOutput atomic.Value

	quit     bool
	listener net.Listener
	t        *testing.T
}

// NewSSHServer returns a NewSSHServer instance, ready for use.
func NewSSHServer(t *testing.T) (*SSHServer, error) {
	t.Helper()
	s := &SSHServer{
		Transfers: &bytes.Buffer{},
		Config:    &ssh.ServerConfig{NoClientAuth: true},
		Commands:  map[string]int{},
		t:         t,
	}

	private, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, errors.Wrap(err, "Error generating RSA key")
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating signer from key")
	}
	s.Config.AddHostKey(signer)
	s.SetSessionRequested(false)
	s.SetCommandToOutput(map[string]string{})
	return s, nil
}

type execRequest struct {
	Command string
}

// Serve loop, listen for connections and store the commands.
func (s *SSHServer) serve() {
	s.t.Logf("Serving ...")
	loop := 0
	for {
		loop++
		s.t.Logf("[loop %d] Accepting for %v...", loop, s)
		c, err := s.listener.Accept()
		if s.quit {
			return
		}
		if err != nil {
			s.t.Errorf("Listener: %v", err)
			return
		}
		go s.handleIncomingConnection(c)
	}
}

// handle an incoming ssh connection
func (s *SSHServer) handleIncomingConnection(c net.Conn) {
	var wg sync.WaitGroup

	_, chans, reqs, err := ssh.NewServerConn(c, s.Config)
	if err != nil {
		s.t.Logf("newserverconn error: %v", err)
		return
	}
	// The incoming Request channel must be serviced.
	wg.Add(1)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
	}()

	// Service the incoming Channel channel.
	for newChannel := range chans {
		if newChannel.ChannelType() == "session" {
			s.SetSessionRequested(true)
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			s.t.Logf("ch accept err: %v", err)
			return
		}
		s.Connected = true
		for req := range requests {
			s.handleRequest(channel, req, &wg)
		}
	}
	wg.Wait()
}

func (s *SSHServer) handleRequest(channel ssh.Channel, req *ssh.Request, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		// Explicitly copy buffer contents to avoid data race
		b := s.Transfers.Bytes()
		if _, err := io.Copy(bytes.NewBuffer(b), channel); err != nil {
			s.t.Errorf("copy failed: %v", err)
		}
		channel.Close()
		wg.Done()
	}()

	switch req.Type {
	case "exec":
		s.t.Logf("exec request received: %+v", req)
		if err := req.Reply(true, nil); err != nil {
			s.t.Errorf("reply failed: %v", err)
		}

		// Note: string(req.Payload) adds additional characters to start of input.
		var cmd execRequest
		if err := ssh.Unmarshal(req.Payload, &cmd); err != nil {
			s.t.Errorf("unmarshal failed: %v", err)
		}
		s.Commands[cmd.Command] = 1

		s.t.Logf("returning output for %s ...", cmd.Command)
		// Write specified command output as mocked ssh output
		if val, err := s.GetCommandToOutput(cmd.Command); err == nil {
			if _, err := channel.Write([]byte(val)); err != nil {
				s.t.Errorf("Write failed: %v", err)
			}
		}

		s.t.Logf("setting exit-status for %s ...", cmd.Command)
		if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0}); err != nil {
			s.t.Errorf("SendRequest failed: %v", err)
		}

	case "pty-req":
		s.t.Logf("pty request received: %+v", req)
		if err := req.Reply(true, nil); err != nil {
			s.t.Errorf("Reply failed: %v", err)
		}

		if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0}); err != nil {
			s.t.Errorf("SendRequest failed: %v", err)
		}
	}
}

// Start the mock SSH Server
func (s *SSHServer) Start() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, errors.Wrap(err, "Error creating tcp listener for ssh server")
	}
	s.listener = l
	s.t.Logf("Listening on %s", s.listener.Addr())
	go s.serve()

	_, p, err := net.SplitHostPort(s.listener.Addr().String())
	if err != nil {
		return 0, errors.Wrap(err, "Error splitting host port")
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, errors.Wrap(err, "Error converting port string to integer")
	}
	return port, nil
}

// Stop the mock SSH server
func (s *SSHServer) Stop() {
	s.t.Logf("Stopping")
	s.quit = true
	s.listener.Close()
}

// SetCommandToOutput sets command to output
func (s *SSHServer) SetCommandToOutput(cmdToOutput map[string]string) {
	s.commandToOutput.Store(cmdToOutput)
}

// GetCommandToOutput gets command to output
func (s *SSHServer) GetCommandToOutput(cmd string) (string, error) {
	cmdMap := s.commandToOutput.Load().(map[string]string)
	val, ok := cmdMap[cmd]
	if !ok {
		return "", fmt.Errorf("unavailable command %s", cmd)
	}
	return val, nil
}

// SetSessionRequested sets session requested
func (s *SSHServer) SetSessionRequested(b bool) {
	var i int32
	if b {
		i = 1
	}
	atomic.StoreInt32(&s.hadASessionRequested, i)
}

// IsSessionRequested gcode ets session requested
func (s *SSHServer) IsSessionRequested() bool {
	return atomic.LoadInt32(&s.hadASessionRequested) != 0
}
