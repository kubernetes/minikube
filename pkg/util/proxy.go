/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package util

import (
	"io"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang/glog"
)

// Proxy will listen at the given address on the "tcp" network
// and forward all incoming connections to the host and port
// specified by the url. It rewrites the URL so that it
// refers to the proxy and returns that.
//
// Forwarding is active as long as the program runs.
func Proxy(listenAddress string, targetURL string) (proxyURL string, err error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	// Start listening.
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return "", err
	}

	// Keep listening for incoming connections in the background
	// and forward the data both ways.
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				glog.Fatalf("Failed to accept connection: %v", err)
			}
			go func() {
				client, err := net.Dial("tcp", host+":"+port)
				if err != nil {
					glog.Fatalf("Failed to connect: %v", err)
				}
				go func() {
					defer client.Close()
					defer conn.Close()
					io.Copy(client, conn)
				}()
				go func() {
					defer client.Close()
					defer conn.Close()
					io.Copy(conn, client)
				}()
			}()
		}
	}()

	// Rewrite URL.
	addr := listener.Addr().String()
	// Here we have to rely on the observation that Addr has [::] as prefix when Listen
	// was called without a specific address. This is not actually documented.
	if strings.HasPrefix(addr, "[::]:") {
		// Just a port number. For the URL to be useful outside of the local machine, we have to add
		// some kind of hostname. A very elaborate scheme would look up local IP addresses, but
		// simply picking the hostname should give reasonable results in a local LAN and is much simpler.
		addr = addr[4:]
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "localhost"
		}
		addr = hostname + addr
	}
	u.Host = addr
	return u.String(), nil
}

// RunTillBreak sleeps until interupted by SIGINT/TERM. Useful in
// a program which has set up port forwarding with Proxy.
func RunTillBreak() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}
