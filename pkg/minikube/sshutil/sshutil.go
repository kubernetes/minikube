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
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	machinessh "github.com/docker/machine/libmachine/ssh"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/util/retry"
)

// NewSSHClient returns an SSH client object for running commands.
func NewSSHClient(d drivers.Driver) (*ssh.Client, error) {
	h, err := newSSHHost(d)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new ssh host from driver")

	}
	defaultKeyPath := filepath.Join(homedir.HomeDir(), ".ssh", "id_rsa")
	auth := &machinessh.Auth{}
	if h.SSHKeyPath != "" {
		auth.Keys = []string{h.SSHKeyPath}
	} else {
		auth.Keys = []string{defaultKeyPath}
	}

	klog.Infof("new ssh client: %+v", h)

	config, err := machinessh.NewNativeConfig(h.Username, auth)
	if err != nil {
		return nil, errors.Wrapf(err, "Error creating new native config from ssh using: %s, %s", h.Username, auth)
	}

	var client *ssh.Client
	getSSH := func() (err error) {
		client, err = ssh.Dial("tcp", net.JoinHostPort(h.IP, strconv.Itoa(h.Port)), &config)
		if err != nil {
			klog.Warningf("dial failure (will retry): %v", err)
		}
		return err
	}

	if err := retry.Expo(getSSH, 250*time.Millisecond, 2*time.Second); err != nil {
		return nil, err
	}

	return client, nil
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

	// For KIC driver with remote Docker context, establish SSH tunnel if needed
	// TODO: Re-enable this once EstablishSSHTunnelForContainer is implemented
	// if driver.IsKIC(d.DriverName()) && oci.IsRemoteDockerContext() && oci.IsSSHDockerContext() {
	// 	// Extract container name from machine name
	// 	containerName := d.GetMachineName()
	// 	klog.Infof("Establishing SSH tunnel for remote Docker container: %s", containerName)
	//
	// 	tunnelPort, err := oci.EstablishSSHTunnelForContainer(containerName, 22)
	// 	if err != nil {
	// 		klog.Warningf("Failed to establish SSH tunnel for container %s: %v", containerName, err)
	// 		// Fall back to direct connection
	// 	} else {
	// 		klog.Infof("SSH tunnel established for container %s: localhost:%d -> remote:%d", containerName, tunnelPort, port)
	// 		// Use tunnel endpoint
	// 		ip = "127.0.0.1"
	// 		port = tunnelPort
	// 	}
	// }

	return &sshHost{
		IP:         ip,
		Port:       port,
		SSHKeyPath: d.GetSSHKeyPath(),
		Username:   d.GetSSHUsername(),
	}, nil
}

// KnownHost checks if this host is in the knownHosts file
func KnownHost(host string, knownHosts string) bool {
	fd, err := os.Open(knownHosts)
	if err != nil {
		return false
	}
	defer fd.Close()

	hashhost := knownhosts.HashHostname(host)
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		_, hosts, _, _, _, err := ssh.ParseKnownHosts(scanner.Bytes())
		if err != nil {
			continue
		}

		for _, h := range hosts {
			if h == host || h == hashhost {
				return true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		klog.Warningf("failed to read file: %v", err)
	}

	return false
}
