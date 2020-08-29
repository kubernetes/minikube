/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

var (
	checkIfRunning func(pid int) (bool, error)
	getPid         func() int
)

func init() {
	checkIfRunning = osCheckIfRunning
	getPid = osGetPid
}

func osGetPid() int {
	return os.Getpid()
}

// TODO(balintp): this is vulnerable to pid reuse we should include process name in the check
func osCheckIfRunning(pid int) (bool, error) {
	p, err := os.FindProcess(pid)
	if runtime.GOOS == "windows" {
		return err == nil, nil
	}
	// on unix systems further checking is required, as findProcess is noop
	if err != nil {
		return false, fmt.Errorf("error finding process %d: %s", pid, err)
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}
	if p == nil {
		return false, nil
	}
	return true, nil
}
