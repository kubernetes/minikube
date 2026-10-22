/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package drivers

import (
	"errors"
	"path/filepath"

	"k8s.io/minikube/pkg/minikube/runtimedir"
)

const (
	DefaultSSHUser          = "root"
	DefaultSSHPort          = 22
	DefaultEngineInstallURL = "https://get.docker.com"
)

// BaseDriver - Embed this struct into drivers to provide the common set
// of fields and functions.
type BaseDriver struct {
	IPAddress      string
	MachineName    string
	SSHUser        string
	SSHPort        int
	SSHKeyPath     string
	StorePath      string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string

	// socketDir is the per-machine socket directory, resolved and created on
	// first use. It is not persisted: the runtime directory may not survive a
	// reboot.
	socketDir string
}

// DriverName returns the name of the driver
func (d *BaseDriver) DriverName() string {
	return "unknown"
}

// GetMachineName returns the machine name
func (d *BaseDriver) GetMachineName() string {
	return d.MachineName
}

// GetIP returns the ip
func (d *BaseDriver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", errors.New("IP address is not set")
	}
	return d.IPAddress, nil
}

// GetSSHKeyPath returns the ssh key path
func (d *BaseDriver) GetSSHKeyPath() string {
	if d.SSHKeyPath == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

// GetSSHPort returns the ssh port, 22 if not specified
func (d *BaseDriver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = DefaultSSHPort
	}

	return d.SSHPort, nil
}

// GetSSHUsername returns the ssh user name, root if not specified
func (d *BaseDriver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = DefaultSSHUser
	}
	return d.SSHUser
}

// PreCreateCheck is called to enforce pre-creation steps
func (d *BaseDriver) PreCreateCheck() error {
	return nil
}

// ResolveStorePath returns the store path where the machine is
func (d *BaseDriver) ResolveStorePath(file string) string {
	return filepath.Join(d.StorePath, "machines", d.MachineName, file)
}

// SocketDir returns the directory holding the AF_UNIX sockets of the machine,
// creating it on first use. Sockets cannot live in the store path, since it
// may be too long for sockaddr_un.sun_path. See pkg/minikube/runtimedir.
func (d *BaseDriver) SocketDir() (string, error) {
	if d.socketDir == "" {
		dir, err := runtimedir.EnsureSocketDir(d.MachineName)
		if err != nil {
			return "", err
		}
		d.socketDir = dir
	}
	return d.socketDir, nil
}

// ResolveSocketPath returns the path of a machine socket, creating the socket
// directory on first use. This is the ResolveStorePath of AF_UNIX sockets.
func (d *BaseDriver) ResolveSocketPath(file string) (string, error) {
	dir, err := d.SocketDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, file), nil
}

// SetSwarmConfigFromFlags configures the driver for swarm
func (d *BaseDriver) SetSwarmConfigFromFlags(flags DriverOptions) {
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
}

func EngineInstallURLFlagSet(flags DriverOptions) bool {
	return EngineInstallURLSet(flags.String("engine-install-url"))
}

func EngineInstallURLSet(url string) bool {
	return url != DefaultEngineInstallURL && url != ""
}
