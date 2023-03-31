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
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/phayes/freeport"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

type sshConn struct {
	name           string
	service        string
	cmd            *exec.Cmd
	ports          []int
	activeConn     bool
	suppressStdOut bool
}

func createSSHConn(name, sshPort, sshKey, bindAddress string, resourcePorts []int32, resourceIP string, resourceName string) *sshConn {
	// extract sshArgs
	sshArgs := []string{
		// TODO: document the options here
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
		"-o", "IdentitiesOnly=yes",
		"-N",
		"docker@127.0.0.1",
		"-p", sshPort,
		"-i", sshKey,
	}

	askForSudo := false
	var privilegedPorts []int32
	for _, port := range resourcePorts {
		var arg string
		if bindAddress == "" || bindAddress == "*" {
			// bind on all interfaces
			arg = fmt.Sprintf(
				"-L %d:%s:%d",
				port,
				resourceIP,
				port,
			)
		} else {
			// bind on specify address only
			arg = fmt.Sprintf(
				"-L %s:%d:%s:%d",
				bindAddress,
				port,
				resourceIP,
				port,
			)
		}

		// check if any port is privileged
		if port < 1024 {
			privilegedPorts = append(privilegedPorts, port)
			askForSudo = true
		}

		sshArgs = append(sshArgs, arg)
	}

	command := "ssh"
	if askForSudo && runtime.GOOS != "windows" {
		out.Styled(
			style.Warning,
			"The service/ingress {{.resource}} requires privileged ports to be exposed: {{.ports}}",
			out.V{"resource": resourceName, "ports": fmt.Sprintf("%v", privilegedPorts)},
		)

		out.Styled(style.Permissions, "sudo permission will be asked for it.")

		command = "sudo"
		sshArgs = append([]string{"ssh"}, sshArgs...)
	}

	if askForSudo && runtime.GOOS == "windows" {
		out.WarningT("Access to ports below 1024 may fail on Windows with OpenSSH clients older than v8.1. For more information, see: https://minikube.sigs.k8s.io/docs/handbook/accessing/#access-to-ports-1024-on-windows-requires-root-permission")
	}

	cmd := exec.Command(command, sshArgs...)

	return &sshConn{
		name:       name,
		service:    resourceName,
		cmd:        cmd,
		activeConn: false,
	}
}

func createSSHConnWithRandomPorts(name, sshPort, sshKey string, svc *v1.Service) (*sshConn, error) {
	// extract sshArgs
	sshArgs := []string{
		// TODO: document the options here
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
		"-o", "IdentitiesOnly=yes",
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
		name:       name,
		service:    svc.Name,
		cmd:        cmd,
		ports:      usedPorts,
		activeConn: false,
	}, nil
}

func (c *sshConn) startAndWait() error {
	if !c.suppressStdOut {
		out.Step(style.Running, "Starting tunnel for service {{.service}}.", out.V{"service": c.service})
	}

	r, w := io.Pipe()
	c.cmd.Stdout = w
	c.cmd.Stderr = w
	err := c.cmd.Start()
	if err != nil {
		return err
	}
	go logOutput(r, c.service)

	c.activeConn = true
	// we ignore wait error because the process will be killed
	_ = c.cmd.Wait()

	// Wait is finished for connection, mark false.
	c.activeConn = false

	return nil
}

func logOutput(r io.Reader, service string) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		klog.Infof("%s tunnel: %s", service, s.Text())
	}
	if err := s.Err(); err != nil {
		klog.Warningf("failed to read: %v", err)
	}
}

func (c *sshConn) stop() error {
	if c.activeConn {
		c.activeConn = false
		if !c.suppressStdOut {
			out.Step(style.Stopping, "Stopping tunnel for service {{.service}}.", out.V{"service": c.service})
		}
		err := c.cmd.Process.Kill()
		if err == os.ErrProcessDone {
			// No need to return an error here
			return nil
		}
		return err
	}
	if !c.suppressStdOut {
		out.Step(style.Stopping, "Stopped tunnel for service {{.service}}.", out.V{"service": c.service})
	}
	return nil
}
