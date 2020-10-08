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
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
)

// CreateSSHShell creates a new SSH shell / client
func CreateSSHShell(api libmachine.API, cc config.ClusterConfig, n config.Node, args []string, native bool) error {
	machineName := driver.MachineName(cc, n)
	host, err := LoadHost(api, machineName)
	if err != nil {
		return errors.Wrap(err, "host exists and load")
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return errors.Wrap(err, "state")
	}

	if currentState != state.Running {
		return errors.Errorf("%q is not running", machineName)
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
