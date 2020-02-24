package kic

import (
	"fmt"
	"os/exec"

	v1 "k8s.io/api/core/v1"
)

type sshConn struct {
	name string
	cmd  *exec.Cmd
}

func createSSHConn(name, sshPort, sshKey string, svc v1.Service) *sshConn {
	// extract sshArgs
	sshArgs := []string{
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

	cmd := exec.Command("ssh", sshArgs...)

	return &sshConn{
		name: name,
		cmd:  cmd,
	}
}

func (c *sshConn) run() error {
	fmt.Printf("starting tunnel for %s\n", c.name)
	return c.cmd.Run()
}

func (c *sshConn) stop() error {
	fmt.Printf("stopping tunnel for %s\n", c.name)
	return c.cmd.Process.Kill()
}
