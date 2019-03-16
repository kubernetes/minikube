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
	"sync/atomic"

	"github.com/golang/glog"
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
	// commandsToOutput can be used to mock what the SSHServer returns for a given command
	// Only access this with atomic ops
	commandToOutput atomic.Value
}

// NewSSHServer returns a NewSSHServer instance, ready for use.
func NewSSHServer() (*SSHServer, error) {
	s := &SSHServer{}
	s.Transfers = &bytes.Buffer{}
	s.Config = &ssh.ServerConfig{
		NoClientAuth: true,
	}
	s.Commands = make(map[string]int)

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

// Start starts the mock SSH Server, and returns the port it's listening on.
func (s *SSHServer) Start() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, errors.Wrap(err, "Error creating tcp listener for ssh server")
	}

	// Main loop, listen for connections and store the commands.
	go func() {
		for {
			nConn, err := listener.Accept()
			go func() {
				if err != nil {
					return
				}

				_, chans, reqs, err := ssh.NewServerConn(nConn, s.Config)
				if err != nil {
					return
				}
				// The incoming Request channel must be serviced.
				go ssh.DiscardRequests(reqs)

				// Service the incoming Channel channel.
				for newChannel := range chans {
					if newChannel.ChannelType() == "session" {
						s.SetSessionRequested(true)
					}
					channel, requests, err := newChannel.Accept()
					s.Connected = true
					if err != nil {
						return
					}

					for req := range requests {
						glog.Infoln("Got Req: ", req.Type)
						// Store anything that comes in over stdin.
						go func() {
							io.Copy(s.Transfers, channel)
							channel.Close()
						}()
						switch req.Type {
						case "exec":
							req.Reply(true, nil)

							// Note: string(req.Payload) adds additional characters to start of input.
							var cmd execRequest
							if err := ssh.Unmarshal(req.Payload, &cmd); err != nil {
								glog.Errorf("Unmarshall encountered error: %v with req: %v", err, req.Type)
								return
							}
							s.Commands[cmd.Command] = 1

							// Write specified command output as mocked ssh output
							if val, err := s.GetCommandToOutput(cmd.Command); err == nil {
								channel.Write([]byte(val))
							}
							channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})

						case "pty-req":
							req.Reply(true, nil)

							channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
						}
					}
				}
			}()
		}
	}()

	// Parse and return the port.
	_, p, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return 0, errors.Wrap(err, "Error splitting host port")
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, errors.Wrap(err, "Error converting port string to integer")
	}
	return port, nil
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
