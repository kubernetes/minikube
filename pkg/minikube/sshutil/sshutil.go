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
	"net"
	"strconv"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	machinessh "github.com/docker/machine/libmachine/ssh"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/util/retry"
)

// NewSSHClient returns an SSH client object for running commands.
func NewSSHClient(d drivers.Driver) (*ssh.Client, error) {
	h, err := newSSHHost(d)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new ssh host from driver")

	}
	auth := &machinessh.Auth{}
	if h.SSHKeyPath != "" {
		auth.Keys = []string{h.SSHKeyPath}
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
	return &sshHost{
		IP:         ip,
		Port:       port,
		SSHKeyPath: d.GetSSHKeyPath(),
		Username:   d.GetSSHUsername(),
	}, nil
}
