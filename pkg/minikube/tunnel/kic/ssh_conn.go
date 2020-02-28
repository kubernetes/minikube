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

package kic

import (
	"fmt"
	"os/exec"

	"github.com/phayes/freeport"

	v1 "k8s.io/api/core/v1"
)

type sshConn struct {
	name    string
	service string
	cmd     *exec.Cmd
	ports   []int
}

func createSSHConn(name, sshPort, sshKey string, svc *v1.Service) *sshConn {
	// extract sshArgs
	sshArgs := []string{
		// TODO: document the options here
		"ssh",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking no",
		"-N",
		"docker@127.0.0.1",
		"-p", sshPort,
		"-i", sshKey,
	}

	for _, port := range svc.Spec.Ports {
		arg := fmt.Sprintf(
			"-L %d:%s:%d",
			port.Port,
			svc.Spec.ClusterIP,
			port.Port,
		)

		sshArgs = append(sshArgs, arg)
	}

	cmd := exec.Command("sudo", sshArgs...)

	return &sshConn{
		name:    name,
		service: svc.Name,
		cmd:     cmd,
	}
}

func createSSHConnWithRandomPorts(name, sshPort, sshKey string, svc *v1.Service) (*sshConn, error) {
	// extract sshArgs
	sshArgs := []string{
		// TODO: document the options here
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking no",
		"-N",
		"docker@127.0.0.1",
		"-p", sshPort,
		"-i", sshKey,
	}

	usedPorts := make([]int, 0, len(svc.Spec.Ports))

	for _, port := range svc.Spec.Ports {
		freeport, err := freeport.GetFreePort()
		if err != nil {
			return nil, err
		}

		arg := fmt.Sprintf(
			"-L %d:%s:%d",
			freeport,
			svc.Spec.ClusterIP,
			port.Port,
		)

		sshArgs = append(sshArgs, arg)
		usedPorts = append(usedPorts, freeport)
	}

	cmd := exec.Command("ssh", sshArgs...)

	return &sshConn{
		name:    name,
		service: svc.Name,
		cmd:     cmd,
		ports:   usedPorts,
	}, nil
}

func (c *sshConn) startAndWait() error {
	fmt.Printf("starting tunnel for %s\n", c.service)
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	// we ignore wait error because the process will be killed
	_ = c.cmd.Wait()

	return nil
}

func (c *sshConn) stop() error {
	fmt.Printf("stopping tunnel for %s\n", c.service)
	return c.cmd.Process.Kill()
}
