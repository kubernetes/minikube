/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"os/exec"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
)

func TestFakeRunnerFile(t *testing.T) {
	fakeCommandRunner := NewFakeCommandRunner()
	cmdArg := "test"
	cmdToOutput := make(map[string]string)
	cmdToOutput[cmdArg] = "123"
	fakeCommandRunner.SetCommandToOutput(cmdToOutput)

	t.Run("SetGetFileContents", func(t *testing.T) {
		fileToContentsMap := make(map[string]string)
		fileName := "fileName"
		expectedFileContents := "fileContents"
		fileToContentsMap[fileName] = expectedFileContents

		fakeCommandRunner.SetFileToContents(fileToContentsMap)

		retrievedFileContents, err := fakeCommandRunner.GetFileToContents(fileName)
		if err != nil {
			t.Fatal(err)
		}

		if expectedFileContents != retrievedFileContents {
			t.Errorf("expected %q, retrieved %q", expectedFileContents, retrievedFileContents)
		}
	})

	t.Run("CopyRemoveFile", func(t *testing.T) {
		expectedFileContents := "test contents"
		fileName := "memory"
		file := assets.NewMemoryAssetTarget([]byte(expectedFileContents), "", "")

		if err := fakeCommandRunner.Copy(file); err != nil {
			t.Fatal(err)
		}

		retrievedFileContents, err := fakeCommandRunner.GetFileToContents(fileName)
		if err != nil {
			t.Fatal(err)
		}

		if expectedFileContents != retrievedFileContents {
			t.Errorf("expected %q, retrieved %q", expectedFileContents, retrievedFileContents)
		}

		if err := fakeCommandRunner.Remove(file); err != nil {
			t.Fatal(err)
		}

		if _, err := fakeCommandRunner.GetFileToContents(fileName); err == nil {
			t.Errorf("file was not removed")
		}
	})

	t.Run("RunCmd", func(t *testing.T) {
		expectedOutput := "123"
		command := &exec.Cmd{Args: []string{cmdArg}}

		rr, err := fakeCommandRunner.RunCmd(command)
		if err != nil {
			t.Fatal(err)
		}

		retrievedOutput := rr.Stdout.String()
		if expectedOutput != retrievedOutput {
			t.Errorf("expected %q, retrieved %q", expectedOutput, retrievedOutput)
		}
	})

	t.Run("StartWaitCmd", func(t *testing.T) {
		expectedOutput := "123"
		command := &exec.Cmd{Args: []string{cmdArg}}

		sc, err := fakeCommandRunner.StartCmd(command)
		if err != nil {
			t.Fatal(err)
		}

		retrievedOutput := sc.rr.Stdout.String()
		if expectedOutput != retrievedOutput {
			t.Errorf("expected %q, retrieved %q", expectedOutput, retrievedOutput)
		}

		rr, err := fakeCommandRunner.WaitCmd(sc)
		if err != nil {
			t.Fatal(err)
		}

		retrievedOutput = rr.Stdout.String()
		if expectedOutput != retrievedOutput {
			t.Errorf("expected %q, retrieved %q", expectedOutput, retrievedOutput)
		}

	})
}
