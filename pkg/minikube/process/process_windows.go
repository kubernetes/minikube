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
)

func pidExists(pid int) (bool, error) {
	// Fails with "OpenProcess: The parameter is incorrect" if the process does
	// not exist.
	_, err := os.FindProcess(pid)
	return err == nil, nil
}

func terminatePid(pid int) error {
	return killPid(pid)
}

func killPid(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return os.ErrProcessDone
	}
	return p.Kill()
}
