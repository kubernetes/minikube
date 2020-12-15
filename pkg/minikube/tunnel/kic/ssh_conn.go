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
	"runtime"

	"github.com/phayes/freeport"
	v1 "k8s.io/api/core/v1"

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
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
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
		"-N",
		"docker@127.0.0.1",
		"-p", sshPort,
		"-i", sshKey,
	}

	askForSudo := false
	var privilegedPorts []int32
	for _, port := range svc.Spec.Ports {
		arg := fmt.Sprintf(
			"-L %d:%s:%d",
			port.Port,
			svc.Spec.ClusterIP,
			port.Port,
		)

		// check if any port is privileged
		if port.Port < 1024 {
			privilegedPorts = append(privilegedPorts, port.Port)
			askForSudo = true
		}

		sshArgs = append(sshArgs, arg)
	}

	command := "ssh"
	if askForSudo && runtime.GOOS != "windows" {
		out.Step(
			style.Warning,
			"The service {{.service}} requires privileged ports to be exposed: {{.ports}}",
			out.V{"service": svc.Name, "ports": fmt.Sprintf("%v", privilegedPorts)},
		)

		out.Step(style.Permissions, "sudo permission will be asked for it.")

		command = "sudo"
		sshArgs = append([]string{"ssh"}, sshArgs...)
	}
	cmd := exec.Command(command, sshArgs...)

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
		"-o", "StrictHostKeyChecking=no",
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
	out.Step(style.Running, "Starting tunnel for service {{.service}}.", out.V{"service": c.service})

	err := c.cmd.Start()
	if err != nil {
		return err
	}

	// we ignore wait error because the process will be killed
	_ = c.cmd.Wait()

	return nil
}

func (c *sshConn) stop() error {
	out.Step(style.Stopping, "Stopping tunnel for service {{.service}}.", out.V{"service": c.service})

	return c.cmd.Process.Kill()
}
