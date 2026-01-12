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

package provision

import (
	"fmt"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/ssh"
)

type RedHatSSHCommander struct {
	Driver drivers.Driver
}

func (sshCmder RedHatSSHCommander) SSHCommand(args string) (string, error) {
	client, err := drivers.GetSSHClientFromDriver(sshCmder.Driver)
	if err != nil {
		return "", err
	}

	log.Debugf("About to run SSH command:\n%s", args)

	// redhat needs "-t" for tty allocation on ssh therefore we check for the
	// external client and add as needed.
	// Note: CentOS 7.0 needs multiple "-tt" to force tty allocation when ssh has
	// no local tty.
	var output string
	switch c := client.(type) {
	case *ssh.ExternalClient:
		c.BaseArgs = append(c.BaseArgs, "-tt")
		output, err = c.Output(args)
	case *ssh.NativeClient:
		output, err = c.OutputWithPty(args)
	}

	log.Debugf("SSH cmd err, output: %v: %s", err, output)
	if err != nil {
		return "", fmt.Errorf(`something went wrong running an SSH command
command : %s
err     : %v
output  : %s`, args, err, output)
	}

	return output, nil
}
