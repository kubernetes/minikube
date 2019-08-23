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
	"io"

	"k8s.io/minikube/pkg/minikube/assets"
)

// OciRunner runs commands inside container
type OciRunner struct {
	Profile string
	stdin   io.Reader
}

// NewOciRunner returns a new OciRunner that will run commands
func NewOciRunner(p string) (*OciRunner, error) {
	return &OciRunner{Profile: p}, nil
}

// Remove runs a command to delete a file on the remote.
func (o *OciRunner) Remove(f assets.CopyableFile) error {
	// MEDYA:TODO later
	return nil
}

// Run starts a command on the remote and waits for it to return.
func (o *OciRunner) Run(cmd string) error {
	// MEDYA:TODO later
	return nil
}

// CombinedOutputTo runs the command and stores both command
// output and error to out.
func (o *OciRunner) CombinedOutputTo(cmd string, w io.Writer) error {
	out, err := o.CombinedOutput(cmd)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(out))
	return err
}

// CombinedOutput runs the command on the remote and returns its combined
// standard output and standard error.
func (o *OciRunner) CombinedOutput(cmd string) (string, error) {
	// MEDYA:TODO later
	return "", nil
}

// Copy copies a file to the remote over SSH.
func (o *OciRunner) Copy(f assets.CopyableFile) error {
	// MEDYA:TODO later
	return nil
}

// SetStdin is used in piping commands
func (o *OciRunner) SetStdin(rd io.Reader) Runner {
	o.stdin = rd
	return o
}
