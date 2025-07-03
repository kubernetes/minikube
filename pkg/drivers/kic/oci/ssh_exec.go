/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
	"k8s.io/klog/v2"
)

// CreateSSHTerminal creates an interactive SSH-like terminal to the container
func CreateSSHTerminal(containerName string, args []string) error {
	klog.Warningf("CreateSSHTerminal called for container %s with args %v", containerName, args)
	klog.Warningf("IsRemoteDockerContext(): %v", IsRemoteDockerContext())

	if !IsRemoteDockerContext() {
		// For local Docker, use standard SSH
		klog.Infof("Not using remote Docker context, falling back to standard SSH")
		return nil
	}

	// For remote Docker contexts, use docker exec
	klog.Warningf("Using docker exec for SSH-like access to remote container %s", containerName)

	cmdArgs := []string{"exec"}

	// Only use -it if we have a TTY
	if isatty.IsTerminal(os.Stdout.Fd()) {
		cmdArgs = append(cmdArgs, "-it")
	} else {
		cmdArgs = append(cmdArgs, "-i")
	}

	cmdArgs = append(cmdArgs, containerName)

	if len(args) > 0 {
		// If we have arguments, execute them through bash -c
		cmdArgs = append(cmdArgs, "/bin/bash", "-c")
		cmdArgs = append(cmdArgs, args...)
	} else {
		// Default to bash shell
		cmdArgs = append(cmdArgs, "/bin/bash")
	}

	cmd := exec.Command(Docker, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	klog.Warningf("Executing docker command: %s %v", Docker, cmdArgs)
	return cmd.Run()
}