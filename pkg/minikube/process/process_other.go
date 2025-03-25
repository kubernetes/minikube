//go:build !windows

/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package process

import (
	"os"
	"syscall"
)

func pidExists(pid int) (bool, error) {
	// Never fails and we get a process in "done" state that returns
	// os.ErrProcessDone from Signal or Wait.
	process, err := os.FindProcess(pid)
	if err != nil {
		return true, err
	}
	if process.Signal(syscall.Signal(0)) == os.ErrProcessDone {
		return false, nil
	}
	return true, nil
}

func terminatePid(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGTERM)
}

func killPid(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
