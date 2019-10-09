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
	"fmt"
	"os/exec"
	"path"
	"strings"

	"k8s.io/minikube/pkg/minikube/assets"
)

type RunResult struct {
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
	ExitCode int
	Args     []string
}

// Runner represents an interface to run commands.
type Runner interface {
	// RunCmd is a new expermintal way to run commands, takes Cmd interface and returns run result.
	// if succesfull will cause a clean up to get rid of older methods.
	RunCmd(cmd *exec.Cmd) (*RunResult, error)

	// Run starts the specified command and waits for it to complete.
	Run(cmd string) error

	// Copy is a convenience method that runs a command to copy a file
	Copy(assets.CopyableFile) error

	// Remove is a convenience method that runs a command to remove a file
	Remove(assets.CopyableFile) error
}

func getDeleteFileCommand(f assets.CopyableFile) string {
	return fmt.Sprintf("sudo rm %s", path.Join(f.GetTargetDir(), f.GetTargetName()))
}

// Command returns a human readable command string that does not induce eye fatigue
func (rr RunResult) Command() string {
	var sb strings.Builder
	sb.WriteString(strings.TrimPrefix(rr.Args[0], "../../"))
	for _, a := range rr.Args[1:] {
		if strings.Contains(a, " ") {
			sb.WriteString(fmt.Sprintf(` "%s"`, a))
			continue
		}
		sb.WriteString(fmt.Sprintf(" %s", a))
	}
	return sb.String()
}

// Output returns human-readable output for an execution result
func (rr RunResult) Output() string {
	var sb strings.Builder
	if rr.Stdout.Len() > 0 {
		sb.WriteString(fmt.Sprintf("-- stdout --\n%s\n-- /stdout --", rr.Stdout.Bytes()))
	}
	if rr.Stderr.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n** stderr ** \n%s\n** /stderr **", rr.Stderr.Bytes()))
	}
	return sb.String()
}
