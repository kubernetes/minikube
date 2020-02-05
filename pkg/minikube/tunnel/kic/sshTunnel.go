package kic

import (
	"fmt"
	"os/exec"

	v1 "k8s.io/api/core/v1"
)

type sshTunnel struct {
	name    string
	sigkill bool
	cmd     *exec.Cmd
}

func createSSHTunnel(name, sshPort, sshKey string, svc v1.Service) *sshTunnel {
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
	s := &sshTunnel{
		name:    name,
		sigkill: false,
		cmd:     cmd,
	}

	go s.run()

	return s
}

func (s *sshTunnel) run() {
	fmt.Println("running", s.name)
	err := s.cmd.Start()
	if err != nil {
		// TODO: improve logging
		fmt.Println(err)
	}

	// we are ignoring wait return, because the process will be killed, once the tunnel is not needed.
	err = s.cmd.Wait()
	if err != nil {
		// TODO: improve logging
		fmt.Println(err)
	}
}

func (s *sshTunnel) stop() {
	fmt.Println("stopping", s.name)
	err := s.cmd.Process.Kill()
	if err != nil {
		// TODO: improve logging
		fmt.Println(err)
	}
}
