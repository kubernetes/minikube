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

package command

import (
	"bytes"
	"io"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/assets"
)

// OciRunner runs commands using an oci client such as docker or podman
// It implements the CommandRunner interface.
type OciRunner struct {
	// binaryPath string
}

// Run starts the specified command in a bash shell and waits for it to complete.
func (o *OciRunner) Run(cmd string) error {
	glog.Infoln("Run:", cmd)
	// c := exec.Command("/bin/bash", "-c", cmd)
	// if err := c.Run(); err != nil {
	// 	return errors.Wrapf(err, "running command: %s", cmd)
	// }
	return nil
}

// CombinedOutputTo runs the command and stores both command
// output and error to out.
func (o *OciRunner) CombinedOutputTo(cmd string, out io.Writer) error {
	glog.Infoln("Run with output:", cmd)
	// c := exec.Command("/bin/bash", "-c", cmd)
	// c.Stdout = out
	// c.Stderr = out
	// err := c.Run()
	// if err != nil {
	// 	return errors.Wrapf(err, "running command: %s\n.", cmd)
	// }

	return nil
}

// CombinedOutput runs the command  in a bash shell and returns its
// combined standard output and standard error.
func (o *OciRunner) CombinedOutput(cmd string) (string, error) {
	var b bytes.Buffer
	// err := e.CombinedOutputTo(cmd, &b)
	// if err != nil {
	// 	return "", errors.Wrapf(err, "running command: %s\n output: %s", cmd, b.Bytes())
	// }
	return b.String(), nil

}

// Copy copies a file and its permissions
func (o *OciRunner) Copy(f assets.CopyableFile) error {

	return nil
}

// Remove removes a file
func (o *OciRunner) Remove(f assets.CopyableFile) error {
	return nil
}
