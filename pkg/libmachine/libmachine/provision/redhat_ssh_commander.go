package provision

import (
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
)

type RedHatSSHCommander struct {
	Driver drivers.Driver
}

func (sshCmder RedHatSSHCommander) SSHCommand(args string) (string, error) {
	rr, err := sshCmder.Driver.RunCmd(exec.Command(args))
	return rr.Stdout.String(), err
}

// x7NOTE: do not toss this logic
// @@ -13,33 +11,6 @@ type RedHatSSHCommander struct {
//  }
//  func (sshCmder RedHatSSHCommander) SSHCommand(args string) (string, error) {
// -	client, err := drivers.GetSSHClientFromDriver(sshCmder.Driver)
// -	if err != nil {
// -		return "", err
// -	}
// -
// -	log.Debugf("About to run SSH command:\n%s", args)
// -
// -	// redhat needs "-t" for tty allocation on ssh therefore we check for the
// -	// external client and add as needed.
// -	// Note: CentOS 7.0 needs multiple "-tt" to force tty allocation when ssh has
// -	// no local tty.
// -	var output string
// -	switch c := client.(type) {
// -	case *ssh.ExternalClient:
// -		c.BaseArgs = append(c.BaseArgs, "-tt")
// -		output, err = c.Output(args)
// -	case *ssh.NativeClient:
// -		output, err = c.OutputWithPty(args)
// -	}
// -
// -	log.Debugf("SSH cmd err, output: %v: %s", err, output)
// -	if err != nil {
// -		return "", fmt.Errorf(`something went wrong running an SSH command
// -command : %s
// -err     : %v
// -output  : %s`, args, err, output)
// -	}
// -
// -	return output, nil
