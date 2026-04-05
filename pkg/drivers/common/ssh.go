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

package common

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/log"
)

var (
	retryDelay = time.Second

	// We have 2 cases:
	// - First start: called after a DHCP lease was created. The host is up and
	//   has an IP address. The SSH server is accessible in 10-1000
	//   milliseconds locally, and few seconds in the GitHub macOS runners.
	// - Second start: The DHCP lease is found immediately in DHCP leases
	//   database but the host is not up yet. SSH is accessible in few seconds
	//   locally, and 2-3 minutes in GitHub macOS runners.
	timeout = 5 * time.Minute
)

// WaitForSSHAccess waits until remote SSH server is responding. Returns an
// error if the wait timed out or could not be started.
func WaitForSSHAccess(d drivers.Driver) error {
	ip, err := d.GetIP()
	if err != nil {
		return err
	}
	port, err := d.GetSSHPort()
	if err != nil {
		return err
	}
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	log.Infof("Waiting until SSH server %q is accessible", addr)

	start := time.Now()
	deadline := start.Add(timeout)
	dialer := net.Dialer{Deadline: deadline}

	for {
		done, err := checkSSHAccess(&dialer, addr)
		if err != nil {
			return err
		}
		if done {
			log.Infof("SSH server %q is accessible in %.3f seconds", addr, time.Since(start).Seconds())
			return nil
		}
		time.Sleep(retryDelay)
	}
}

// checkSSHAccess performs one check, returning:
// - (true, nil) we accessed the SSH server
// - (false, nil) failed to access the server, need to retry again
// - (false , error) timeout accessing the server, do not retry
func checkSSHAccess(dialer *net.Dialer, addr string) (bool, error) {
	if time.Now().After(dialer.Deadline) {
		return false, fmt.Errorf("timeout waiting for SSH server %q", addr)
	}

	log.Debugf("Dialing to SSH server %q", addr)
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return false, fmt.Errorf("timeout dialing to SSH server %q", addr)
		}
		log.Debugf("Failed to dial: %v", err)
		return false, nil
	}

	defer conn.Close()

	if err := conn.SetReadDeadline(dialer.Deadline); err != nil {
		log.Debugf("Failed to set timeout: %v", err)
		return false, nil
	}

	log.Debugf("Reading from SSH server %q", addr)
	if _, err := conn.Read(make([]byte, 1)); err != nil && err != io.EOF {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return false, fmt.Errorf("timeout reading from SSH server %q", addr)
		}
		log.Debugf("Failed to read: %v", err)
		return false, nil
	}

	return true, nil
}
