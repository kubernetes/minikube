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

package bootstrapper

import (
	"fmt"
	"path/filepath"

	"k8s.io/minikube/pkg/minikube/assets"
)

// CommandRunner represents an interface to run commands.
type CommandRunner interface {
	// Run starts the specified command and waits for it to complete.
	Run(cmd string) error

	// CombinedOutput runs the command and returns its combined standard
	// output and standard error.
	CombinedOutput(cmd string) (string, error)

	// Copy is a convenience method that runs a command to copy a file
	Copy(assets.CopyableFile) error

	//Remove is a convenience method that runs a command to remove a file
	Remove(assets.CopyableFile) error
}

func getDeleteFileCommand(f assets.CopyableFile) string {
	return fmt.Sprintf("sudo rm %s", filepath.Join(f.GetTargetDir(), f.GetTargetName()))
}
