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
	"fmt"
	"io"
	"path"

	"k8s.io/minikube/pkg/minikube/assets"
)

// Cmd abstracts over running a command somewhere, this is useful for testing
type Cmd interface {
	Run() error
	// Each entry should be of the form "key=value"
	SetEnv(...string) Cmd
	SetStdin(io.Reader) Cmd
	SetStdout(io.Writer) Cmd
	SetStderr(io.Writer) Cmd
}

// Runner represents an interface to run commands.
type Runner interface {
	Command(c string) Cmd

	// Run starts the specified command and waits for it to complete.
	Run(c Cmd) error

	// CombinedOutputTo runs the command and stores both command
	// output and error to out. A typical usage is:
	//
	//          var b bytes.Buffer
	//          CombinedOutput(cmd, &b)
	//          fmt.Println(b.Bytes())
	//
	// Or, you can set out to os.Stdout, the command output and
	// error would show on your terminal immediately before you
	// cmd exit. This is useful for a long run command such as
	// continuously print running logs.
	CombinedOutputTo(c Cmd, out io.Writer) error

	// CombinedOutput runs the command and returns its combined standard
	// output and standard error.
	CombinedOutput(c Cmd) (string, error)

	// Copy is a convenience method that runs a command to copy a file
	Copy(assets.CopyableFile) error

	// Remove is a convenience method that runs a command to remove a file
	Remove(assets.CopyableFile) error
}

func getDeleteFileCommand(f assets.CopyableFile) string {
	return fmt.Sprintf("sudo rm %s", path.Join(f.GetTargetDir(), f.GetTargetName()))
}
