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

package machine

import (
	"fmt"
	"os/exec"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
)

func getHost(api libmachine.API, cc config.ClusterConfig, n config.Node) (*host.Host, error) {
	machineName := config.MachineName(cc, n)
	host, err := LoadHost(api, machineName)
	if err != nil {
		return nil, errors.Wrap(err, "host exists and load")
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "state")
	}

	if currentState != state.Running {
		return nil, errors.Errorf("%q is not running", machineName)
	}

	return host, nil
}

// CreateSSHShell creates a new SSH shell / client
func CreateSSHShell(api libmachine.API, cc config.ClusterConfig, n config.Node, args []string, native bool) error {
	host, err := getHost(api, cc, n)
	if err != nil {
		return err
	}

	if native {
		ssh.SetDefaultClient(ssh.Native)
	} else {
		ssh.SetDefaultClient(ssh.External)
	}

	client, err := host.CreateSSHClient()

	if err != nil {
		return errors.Wrap(err, "Creating ssh client")
	}
	return client.Shell(args...)
}

func GetSSHHostAddrPort(api libmachine.API, cc config.ClusterConfig, n config.Node) (string, int, error) {
	host, err := getHost(api, cc, n)
	if err != nil {
		return "", 0, err
	}

	addr, err := host.Driver.GetSSHHostname()
	if err != nil {
		return "", 0, err
	}
	port, err := host.Driver.GetSSHPort()
	if err != nil {
		return "", 0, err
	}

	return addr, port, nil
}

// RunSSHHostCommand runs a command to the SSH host
func RunSSHHostCommand(api libmachine.API, cc config.ClusterConfig, n config.Node, command string, args []string) (string, error) {
	addr, port, err := GetSSHHostAddrPort(api, cc, n)
	if err != nil {
		return "", err
	}

	cmdPath, err := exec.LookPath(command)
	if err != nil {
		return "", err
	}

	args = append(args, "-p")
	args = append(args, fmt.Sprintf("%d", port))

	args = append(args, addr)

	cmd := exec.Command(cmdPath, args...)
	output, err := cmd.Output()
	return string(output), err
}
