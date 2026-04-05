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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// re-implementation of private function in https://github.com/golang/go/blob/master/src/syscall/syscall_windows.go#L945
func getProcessEntry(pid int) (pe *syscall.ProcessEntry32, err error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(syscall.Handle(snapshot))

	var processEntry syscall.ProcessEntry32
	processEntry.Size = uint32(unsafe.Sizeof(processEntry))
	err = syscall.Process32First(snapshot, &processEntry)
	if err != nil {
		return nil, err
	}

	for {
		if processEntry.ProcessID == uint32(pid) {
			pe = &processEntry
			return
		}

		err = syscall.Process32Next(snapshot, &processEntry)
		if err != nil {
			return nil, err
		}
	}
}

// getNameAndItsPpid returns the exe file name its parent process id.
func getNameAndItsPpid(pid int) (exefile string, parentid int, err error) {
	pe, err := getProcessEntry(pid)
	if err != nil {
		return "", 0, err
	}

	name := syscall.UTF16ToString(pe.ExeFile[:])
	return name, int(pe.ParentProcessID), nil
}

func Detect() (string, error) {
	shell := os.Getenv("SHELL")

	// if you spawn a Powershell instance from CMD, sometimes the SHELL environment variable still points to CMD in the Powershell instance
	// so if SHELL is pointing to CMD, let's do extra work to get the correct shell
	if shell == "" || filepath.Base(shell) == "cmd.exe" {
		shell, shellppid, err := getNameAndItsPpid(os.Getppid())
		if err != nil {
			return "cmd", err // defaulting to cmd
		}
		shellMapping := mapShell(shell)
		if shellMapping != "" {
			return shellMapping, nil
		} else {
			shell, _, err := getNameAndItsPpid(shellppid)
			if err != nil {
				return "cmd", err // defaulting to cmd
			}
			shellMapping = mapShell(shell)
			if shellMapping != "" {
				return shellMapping, nil
			} else {
				fmt.Printf("You can further specify your shell with either 'cmd' or 'powershell' with the --shell flag.\n\n")
				return "cmd", nil // this could be either powershell or cmd, defaulting to cmd
			}
		}
	}

	if os.Getenv("__fish_bin_dir") != "" {
		return "fish", nil
	}

	if runtime.GOOS == "windows" {
		shell = strings.TrimSuffix(shell, filepath.Ext(shell))
	}

	return filepath.Base(shell), nil
}

func mapShell(shell string) string {
	mappings := map[string]string{
		"cmd":        "cmd",
		"powershell": "powershell",
		"pwsh":       "powershell",
	}
	for k, v := range mappings {
		if strings.Contains(strings.ToLower(shell), k) {
			return v
		}
	}
	return ""
}
