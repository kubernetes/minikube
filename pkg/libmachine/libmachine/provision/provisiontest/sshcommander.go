/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

// Package provisiontest provides utilities for testing provisioners
package provisiontest

import (
	"bytes"
	"errors"
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
)

// x7TODO: take this out of here.. put it inside the libmachine/runner code

// FakeSSHCommanderOptions is intended to create a FakeSSHCommander without actually knowing the underlying sshcommands by passing it to NewSSHCommander
type FakeSSHCommanderOptions struct {
	// Result of the ssh command to look up the FilesystemType
	FilesystemType string
}

// FakeSSHCommander is an implementation of provision.SSHCommander to provide predictable responses set by testing code
// Extend it when needed
type FakeSSHCommander struct {
	Responses map[string]string
}

// NewFakeSSHCommander creates a FakeSSHCommander without actually knowing the underlying sshcommands
func NewFakeSSHCommander(options FakeSSHCommanderOptions) *FakeSSHCommander {
	if options.FilesystemType == "" {
		options.FilesystemType = "ext4"
	}
	sshCmder := &FakeSSHCommander{
		Responses: map[string]string{
			"stat -f -c %T /var/lib": options.FilesystemType + "\n",
		},
	}

	return sshCmder
}

// SSHCommand is an implementation of provision.SSHCommander.SSHCommand to provide predictable responses set by testing code
func (sshCmder *FakeSSHCommander) RunCmd(cmd *exec.Cmd) (*runner.RunResult, error) {
	response, commandRegistered := sshCmder.Responses[cmd.String()]
	if !commandRegistered {
		return nil, errors.New("command not registered in FakeSSHCommander")
	}
	return &runner.RunResult{Stdout: *bytes.NewBuffer([]byte(response))}, nil
}
