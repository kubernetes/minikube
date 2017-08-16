/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
