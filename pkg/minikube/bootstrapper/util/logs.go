package util

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/sshutil"
)

func GetLogsGeneric(c *ssh.Client, follow bool, logsCommand, driver string) (string, error) {
	sess, err := c.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	if follow {
		err := sshutil.GetShell(sess, logsCommand)
		return "", errors.Wrap(err, "error getting shell")
	}
	s, err := cluster.RunCommand(c, driver, logsCommand, false)
	if err != nil {
		return "", errors.Wrap(err, "error running logs command")
	}
	return s, nil
}
