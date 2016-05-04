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

package localkube

import (
	"fmt"
	"io"
	"time"

	kuberest "k8s.io/kubernetes/pkg/client/restclient"
	kubeclient "k8s.io/kubernetes/pkg/client/unversioned"
)

// Server represents a component that Kubernetes depends on. It allows for the management of
// the lifecycle of the component.
type Server interface {
	// Start immediately starts the component.
	Start()
	// Stop begins the process of stopping the components.
	Stop()

	// Name returns a unique identifier for the component.
	Name() string

	// Status provides the state of the server.
	Status() Status
}

// Servers allows operations to be performed on many servers at once.
// Uses slice to preserve ordering.
type Servers []Server

// Get returns a server matching name, returns nil if server doesn't exit.
func (servers Servers) Get(name string) (Server, error) {
	for _, server := range servers {
		if server.Name() == name {
			return server, nil
		}
	}
	return nil, fmt.Errorf("server '%s' does not exist", name)
}

// StartAll starts all services, starting from 0th item and ascending.
func (servers Servers) StartAll() {
	for _, server := range servers {
		fmt.Printf("Starting %s...\n", server.Name())
		server.Start()
	}
}

// StopAll stops all services, starting with the last item.
func (servers Servers) StopAll() {
	for i := len(servers) - 1; i >= 0; i-- {
		server := servers[i]
		fmt.Printf("Stopping %s...\n", server.Name())
		server.Stop()
	}
}

// Start is a helper method to start the Server specified, returns error if server doesn't exist.
func (servers Servers) Start(serverName string) error {
	server, err := servers.Get(serverName)
	if err != nil {
		return err
	}

	server.Start()
	return nil
}

// Stop is a helper method to start the Server specified, returns error if server doesn't exist.
func (servers Servers) Stop(serverName string) error {
	server, err := servers.Get(serverName)
	if err != nil {
		return err
	}

	server.Stop()
	return nil
}

// Status returns a map with the Server name as the key and it's Status as the value.
func (servers Servers) Status() (statuses map[string]Status) {
	for _, server := range servers {
		statuses[server.Name()] = server.Status()
	}
	return statuses
}

// SimpleServer provides a minimal implementation of Server.
type SimpleServer struct {
	ComponentName string
	StartupFn     func()
	ShutdownFn    func()
	StatusFn      func() Status
}

// NoShutdown sets the ShutdownFn to print an error when the server gets shutdown. It returns itself to be chainable.
func (s SimpleServer) NoShutdown() *SimpleServer {
	s.ShutdownFn = func() {
		fmt.Printf("The server '%s' is unstoppable.\n", s.ComponentName)
	}
	return &s
}

// Start calls startup function.
func (s *SimpleServer) Start() {
	s.StartupFn()
}

// Stop calls shutdown function.
func (s *SimpleServer) Stop() {
	if s.ShutdownFn != nil {
		s.ShutdownFn()
	}
}

// Name returns the name of the service.
func (s SimpleServer) Name() string {
	return s.ComponentName
}

// Status calls the status function and returns the the Server's status.
func (s *SimpleServer) Status() Status {
	return s.StatusFn()
}

// Status indicates the condition of a Server.
type Status string

const (
	// Stopped indicates the server is not running.
	Stopped Status = "Stopped"

	// Started indicates the server is running.
	Started = "Started"

	// NotImplemented is returned when Status cannot be determined.
	NotImplemented = "NotImplemented"
)

// until endlessly loops the provided function until a message is received on the done channel.
// The function will wait the duration provided in sleep between function calls. Errors will be sent on provider Writer.
func until(fn func() error, w io.Writer, name string, sleep time.Duration, done <-chan struct{}) {
	var exitErr error
	for {
		select {
		case <-done:
			return
		default:
			exitErr = fn()
			if exitErr == nil {
				fmt.Fprintf(w, pad("%s: Exited with no errors.\n"), name)
			} else {
				fmt.Fprintf(w, pad("%s: Exit with error: %v"), name, exitErr)
			}

			// wait provided duration before trying again
			time.Sleep(sleep)
		}
	}
}

func pad(str string) string {
	return fmt.Sprint("\n%s\n", str)
}

func kubeClient() *kubeclient.Client {
	config := &kuberest.Config{
		Host: APIServerURL,
	}
	client, err := kubeclient.New(config)
	if err != nil {
		panic(err)
	}
	return client
}
