package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"io/ioutil"
	"net"
	"strconv"

	"golang.org/x/crypto/ssh"
)

// SSHServer provides a mock SSH Server for testing. Commands are stored, not executed.
type SSHServer struct {
	Config *ssh.ServerConfig
	// Commands stores the raw commands executed against the server.
	Commands  []string
	Connected bool
}

// NewSSHServer returns a NewSSHServer instance, ready for use.
func NewSSHServer() (*SSHServer, error) {
	s := &SSHServer{}
	s.Config = &ssh.ServerConfig{
		NoClientAuth: true,
	}

	private, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		return nil, err
	}
	s.Config.AddHostKey(signer)
	return s, nil
}

// Start starts the mock SSH Server, and returns the port it's listening on.
func (s *SSHServer) Start() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	// Main loop, listen for connections and store the commands.
	go func() {
		for {
			nConn, err := listener.Accept()
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
				channel, requests, err := newChannel.Accept()
				s.Connected = true
				if err != nil {
					return
				}

				req := <-requests
				req.Reply(true, nil)
				s.Commands = append(s.Commands, string(req.Payload))
				channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})

				// Discard anything that comes in over stdin.
				io.Copy(ioutil.Discard, channel)
				channel.Close()
			}
		}
	}()

	// Parse and return the port.
	_, p, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	return port, nil
}
