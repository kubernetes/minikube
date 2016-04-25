package provision

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/ssh"
)

type RedHatSSHCommander struct {
	Driver drivers.Driver
}

func (sshCmder RedHatSSHCommander) SSHCommand(args string) (string, error) {
	client, err := drivers.GetSSHClientFromDriver(sshCmder.Driver)
	if err != nil {
		return "", err
	}

	// redhat needs "-t" for tty allocation on ssh therefore we check for the
	// external client and add as needed.
	// Note: CentOS 7.0 needs multiple "-tt" to force tty allocation when ssh has
	// no local tty.
	switch c := client.(type) {
	case *ssh.ExternalClient:
		c.BaseArgs = append(c.BaseArgs, "-tt")
		client = c
	case *ssh.NativeClient:
		return c.OutputWithPty(args)
	}

	return client.Output(args)
}
