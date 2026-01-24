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
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "")

	// Determine what shell we're actually running under
	_, shellppid, err := getNameAndItsPpid(os.Getppid())
	assert.NoError(t, err)
	grandparent, _, err := getNameAndItsPpid(shellppid)
	assert.NoError(t, err)

	shell, err := Detect()
	assert.NoError(t, err)

	// Assert based on actual parent process
	if strings.Contains(strings.ToLower(grandparent), "powershell") || strings.Contains(strings.ToLower(grandparent), "pwsh") {
		assert.Equal(t, "powershell", shell)
	} else {
		assert.Equal(t, "cmd", shell)
	}
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

// isKnownWindowsShell returns true if the given raw process name (path or basename)
// corresponds to a known Windows shell or test runner executable.
func isKnownWindowsShell(raw string) bool {
	norm := strings.ToLower(filepath.Base(strings.TrimSpace(raw)))
	switch norm {
	// conhost.exe is the Console Window Host on Windows and often appears
	// between the terminal (cmd/powershell) and the process tree we inspect.
	// Include it because some environments (Windows Terminal, legacy consoles)
	// will report conhost.exe as the immediate ancestor rather than the shell.
	case "powershell.exe", "pwsh.exe", "cmd.exe", "conhost.exe", "go.exe", "bash.exe", "sh.exe", "wsl.exe":
		return true
	default:
		return false
	}
}

func TestGetNameAndItsPpidOfGrandParent(t *testing.T) {
	_, shellppid, err := getNameAndItsPpid(os.Getppid())
	assert.NoError(t, err)
	shell, _, err := getNameAndItsPpid(shellppid)
	assert.NoError(t, err)

	assert.True(t, isKnownWindowsShell(shell), "unexpected grandparent process: raw=%q", shell)
}
