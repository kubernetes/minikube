/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

// Part of this code is heavily inspired/copied by the following file:
// github.com/docker/machine/commands/env.go

package daemonenv

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/golang/glog"
)

// TryConnectivity will try to connect to daemon env from user's POV to detect the problem if it needs reset or not
func TryConnectivity(bin string, ec DockerEnvConfig) ([]byte, error) {
	switch bin {
	case "docker":
		c := exec.Command(bin, "version", "--format={{.Server}}")
		c.Env = append(os.Environ(), dockerEnvVarsList(ec)...)
		glog.Infof("Testing Docker connectivity with: %v", c)
		return c.CombinedOutput()
	default:
		msg := fmt.Sprintf("Tried to test connectivity of unsupported daemon: %s", bin)
		glog.Infof(msg)
		return []byte{}, fmt.Errorf("tried to test connectivity of unsupported daemon: %s", bin)
	}
}

// CreateExternalSSHClient creates external ssh client
func CreateExternalSSHClient(d drivers.Driver) (*ssh.ExternalClient, error) {
	sshBinaryPath, err := exec.LookPath("ssh")
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	addr, err := d.GetSSHHostname()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{}
	if d.GetSSHKeyPath() != "" {
		auth.Keys = []string{d.GetSSHKeyPath()}
	}

	return ssh.NewExternalClient(sshBinaryPath, d.GetSSHUsername(), addr, port, auth)
}

// GetNoProxyVar gets the no_proxy var
func GetNoProxyVar() (string, string) {
	// first check for an existing lower case no_proxy var
	noProxyVar := "no_proxy"
	noProxyValue := os.Getenv("no_proxy")

	// otherwise default to allcaps HTTP_PROXY
	if noProxyValue == "" {
		noProxyVar = "NO_PROXY"
		noProxyValue = os.Getenv("NO_PROXY")
	}
	return noProxyVar, noProxyValue
}
