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

package drivers

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/docker/machine/libmachine/log"
)

func WritePidfile(pidfile string, pid int) error {
	data := fmt.Sprintf("%v", pid)
	return os.WriteFile(pidfile, []byte(data), 0600)
}

func ReadPidfile(pidfile string) (int, error) {
	data, err := os.ReadFile(pidfile)
	if err != nil {
		return -1, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return -1, err
	}
	return pid, nil
}

func SignalPidfile(pidfile string, sig syscall.Signal) error {
	pid, err := ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// Already stopped.
		os.Remove(pidfile)
		return nil
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	log.Infof("Sending signal %q to pid %v", sig, pid)
	if err := process.Signal(sig); err != nil {
		if err != os.ErrProcessDone {
			return err
		}
		// Process done.
		os.Remove(pidfile)
		return nil
	}
	return nil
}

func CheckPid(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.Signal(0))
}
