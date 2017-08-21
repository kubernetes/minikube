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
	"bytes"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/minikube/assets"
)

// FakeCommandRunner mocks command output without running the Commands
//
// It implements the CommandRunner interface and is used for testing.
type FakeCommandRunner struct {
	commandToOutput atomic.Value
	cmdMap          map[string]string

	fileToContents atomic.Value
	fileMap        map[string]string
}

// NewFakeCommandRunner returns a new FakeCommandRunner
//
// The expected output of commands should be set with SetCommandToOutput
func NewFakeCommandRunner() *FakeCommandRunner {
	f := &FakeCommandRunner{
		cmdMap:  make(map[string]string),
		fileMap: make(map[string]string),
	}

	f.SetCommandToOutput(f.cmdMap)
	f.SetFileToContents(f.fileMap)
	return f
}

// Run returns nil if output has been set for the given command text.
func (f *FakeCommandRunner) Run(cmd string) error {
	_, err := f.GetCommandToOutput(cmd)
	return err
}

// CombinedOutput returns the set output for a given command text.
func (f *FakeCommandRunner) CombinedOutput(cmd string) (string, error) {
	return f.GetCommandToOutput(cmd)
}

// Copy adds the filename, file contents key value pair to the stored map.
func (f *FakeCommandRunner) Copy(file assets.CopyableFile) error {
	fileMap := f.fileToContents.Load().(map[string]string)
	var b bytes.Buffer
	_, err := io.Copy(&b, file)
	if err != nil {
		return errors.Wrapf(err, "error reading file: %+v", file)
	}
	fileMap[file.GetAssetName()] = b.String()
	f.SetFileToContents(fileMap)
	return nil
}

// Remove removes the filename, file contents key value pair from the stored map
func (f *FakeCommandRunner) Remove(file assets.CopyableFile) error {
	fileMap := f.fileToContents.Load().(map[string]string)
	delete(fileMap, file.GetAssetName())
	f.SetFileToContents(fileMap)
	return nil
}

// GetCommandToOutput retrieves the stored output for a given command from the stored command map
func (f *FakeCommandRunner) GetCommandToOutput(cmd string) (string, error) {
	cmdMap := f.commandToOutput.Load().(map[string]string)
	val, ok := cmdMap[cmd]
	if !ok {
		return "", fmt.Errorf("unavailable command: %s", cmd)
	}
	return val, nil
}

// SetCommandToOutput stores the command to output map for the FakeCommandRunner
func (f *FakeCommandRunner) SetCommandToOutput(cmdToOutput map[string]string) {
	f.commandToOutput.Store(cmdToOutput)
}

// GetFileToContents returns the value for the stored filename key
func (f *FakeCommandRunner) GetFileToContents(fpath string) (string, error) {
	fileMap := f.fileToContents.Load().(map[string]string)
	val, ok := fileMap[fpath]
	if !ok {
		return "", fmt.Errorf("unavailable file: %+v", fpath)
	}
	return val, nil
}

// SetFileToContents stores the file to contents map for the FakeCommandRunner
func (f *FakeCommandRunner) SetFileToContents(fileToContents map[string]string) {
	f.fileToContents.Store(fileToContents)
}

// DumpMaps prints out the list of stored commands and stored filenames.
func (f *FakeCommandRunner) DumpMaps(w io.Writer) {
	fmt.Fprint(w, "Commands: \n", f.cmdMap)
	fmt.Fprintln(w, "Filenames: ")
	for k := range f.fileMap {
		fmt.Fprintln(w, k)
	}
}
