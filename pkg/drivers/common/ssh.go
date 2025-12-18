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
	"io"
	"net"
	"strconv"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
)

var (
	retryDelay = time.Second
)

// WaitForSSHAccess waits until remote SSH server is responing.
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

	for {
		log.Debugf("Dialing to SSH server %q", addr)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Debugf("Failed to dial: %v", err)
			time.Sleep(retryDelay)
			continue
		}

		defer conn.Close()
		log.Debugf("Reading from SSH server %q", addr)
		if _, err := conn.Read(make([]byte, 1)); err != nil && err != io.EOF {
			log.Debugf("Failed to read: %v", err)
			time.Sleep(retryDelay)
			continue
		}

		log.Infof("SSH server %q is accessible in %.3f seconds", addr, time.Since(start).Seconds())
		return nil
	}
}
