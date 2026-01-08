/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "")

	shell, err := Detect()

	assert.Equal(t, "powershell", shell)
	assert.NoError(t, err)
}

func TestDetectOnSSH(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "c:\\windows\\system32\\windowspowershell\\v1.0\\powershell.exe")

	shell, err := Detect()

	assert.Equal(t, "powershell", shell)
	assert.NoError(t, err)
}
func TestGetNameAndItsPpidOfCurrent(t *testing.T) {
	shell, shellppid, err := getNameAndItsPpid(os.Getpid())

	assert.Equal(t, "shell.test.exe", shell)
	assert.Equal(t, os.Getppid(), shellppid)
	assert.NoError(t, err)
}

func TestGetNameAndItsPpidOfParent(t *testing.T) {
	shell, _, err := getNameAndItsPpid(os.Getppid())

	assert.Equal(t, "go.exe", shell)
	assert.NoError(t, err)
}

func TestGetNameAndItsPpidOfGrandParent(t *testing.T) {
	_, shellppid, err := getNameAndItsPpid(os.Getppid())
	assert.NoError(t, err)
	shell, _, err := getNameAndItsPpid(shellppid)
	assert.NoError(t, err)

	assert.Equal(t, "powershell.exe", shell)
	assert.NoError(t, err)
}
