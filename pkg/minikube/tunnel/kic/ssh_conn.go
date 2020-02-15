package kic

import (
	"fmt"
	"os/exec"

	v1 "k8s.io/api/core/v1"
)

type sshConn struct {
	name    string
	sigkill bool
	cmd     *exec.Cmd
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

	// TODO: name must be different, because if a service was changed,
	// we must remove the old process and create the new one
	s := &sshConn{
		name:    name,
		sigkill: false,
		cmd:     cmd,
	}

	// TODO: create should not run
	go s.run()

	return s
}

func (c *sshConn) run() error {
	fmt.Println("running", c.name)
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	// TODO: can we ignore kills and returns not kills?
	// kilss we have to log nonetheless
	err = c.cmd.Wait()
	if err != nil {
		// TODO: improve logging
		fmt.Println(err)
	}

	return nil
}

func (c *sshConn) stop() error {
	fmt.Println("stopping", c.name)
	return c.cmd.Process.Kill()
}
