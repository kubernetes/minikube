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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// WriteSSHProxyConfig writes SSH configuration for accessing container through remote Docker host
func WriteSSHProxyConfig(containerName string, sshPort int) error {
	if !IsRemoteDockerContext() {
		return nil // No proxy needed for local Docker
	}

	ctx, err := GetCurrentContext()
	if err != nil {
		return errors.Wrap(err, "get current context")
	}

	if !ctx.IsSSH {
		return nil // Only SSH contexts need proxy configuration
	}

	// Parse SSH details from context
	sshURL := strings.TrimPrefix(ctx.Host, "ssh://")
	parts := strings.Split(sshURL, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid SSH endpoint format: %s", ctx.Host)
	}

	user := parts[0]
	host := parts[1]

	// Create SSH config directory  
	baseDir := localpath.MiniPath()
	sshDir := filepath.Join(baseDir, "machines", containerName)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return errors.Wrap(err, "create SSH directory")
	}

	// Write SSH config file
	configPath := filepath.Join(sshDir, "config")
	configContent := fmt.Sprintf(`Host %s
    HostName 127.0.0.1
    Port %d
    User docker
    IdentityFile %s
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    ProxyCommand ssh -W %%h:%%p %s@%s
`, containerName, sshPort, filepath.Join(sshDir, "id_rsa"), user, host)

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return errors.Wrap(err, "write SSH config")
	}

	klog.Infof("Wrote SSH proxy config for %s to %s", containerName, configPath)
	return nil
}

// GetSSHCommandForContainer returns the SSH command to connect to a container
func GetSSHCommandForContainer(containerName string) []string {
	baseDir := localpath.MiniPath()
	if !IsRemoteDockerContext() {
		// Local Docker - direct SSH
		sshKey := filepath.Join(baseDir, "machines", containerName, "id_rsa")
		port, _ := ForwardedPort(Docker, containerName, 22)
		return []string{
			"ssh",
			"-i", sshKey,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-p", fmt.Sprintf("%d", port),
			"docker@127.0.0.1",
		}
	}

	// Remote Docker - use SSH config with ProxyCommand
	configPath := filepath.Join(baseDir, "machines", containerName, "config")
	return []string{
		"ssh",
		"-F", configPath,
		containerName,
	}
}

// TestRemoteSSHConnection tests SSH connectivity to container through remote host
func TestRemoteSSHConnection(containerName string) error {
	cmd := GetSSHCommandForContainer(containerName)
	testCmd := exec.Command(cmd[0], append(cmd[1:], "echo", "test")...)
	
	output, err := testCmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "SSH test failed: %s", string(output))
	}
	
	if strings.TrimSpace(string(output)) != "test" {
		return fmt.Errorf("unexpected SSH output: %s", string(output))
	}
	
	klog.Infof("SSH connection to %s successful", containerName)
	return nil
}