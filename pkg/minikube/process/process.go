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
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/go-ps"
)

const pidfileMode = 0o600

// WritePidfile writes pid to path.
func WritePidfile(path string, pid int) error {
	data := fmt.Sprintf("%d", pid)
	return os.WriteFile(path, []byte(data), pidfileMode)
}

// ReadPid reads a pid from path.
func ReadPidfile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Pass os.ErrNotExist
		return -1, err
	}
	s := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(s)
	if err != nil {
		return -1, fmt.Errorf("invalid pid %q: %s", s, err)
	}
	return pid, nil
}

// Exists tells if a process matching pid and executable name exist. Executable is
// not the path to the executable.
func Exists(pid int, executable string) (bool, error) {
	// Fast path if pid does not exist.
	exists, err := pidExists(pid)
	if err != nil {
		return true, err
	}
	if !exists {
		return false, nil
	}

	// Slow path if pid exist, depending on the platform. On windows and darwin
	// this fetch all processes from the krenel and find a process with pid. On
	// linux this reads /proc/pid/stat
	entry, err := ps.FindProcess(pid)
	if err != nil {
		return true, err
	}
	if entry == nil {
		return false, nil
	}
	return entry.Executable() == executable, nil
}

// Terminate a process with pid and matching name. Returns os.ErrProcessDone if
// the process does not exist, or nil if termiation was requested. Caller need
// to wait until the process disappears.
func Terminate(pid int, executable string) error {
	exists, err := Exists(pid, executable)
	if err != nil {
		return err
	}
	if !exists {
		return os.ErrProcessDone
	}
	return terminatePid(pid)
}

// Kill a process with pid matching executable name. Returns os.ErrProcessDone
// if the process does not exist or nil the kill was requested. Caller need to
// wait until the process disappears.
func Kill(pid int, executable string) error {
	exists, err := Exists(pid, executable)
	if err != nil {
		return err
	}
	if !exists {
		return os.ErrProcessDone
	}
	return killPid(pid)
}
