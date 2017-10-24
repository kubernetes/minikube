/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

/*
 * Copyright 2017 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmware

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine/log"
)

var (
	vmrunbin    = setVmwareCmd("vmrun")
	vdiskmanbin = setVmwareCmd("vmware-vdiskmanager")
)

var (
	ErrMachineExist    = errors.New("machine already exists")
	ErrMachineNotExist = errors.New("machine does not exist")
	ErrVMRUNNotFound   = errors.New("VMRUN not found")
)

func init() {
	// vmrun with nogui on VMware Fusion through at least 8.0.1 doesn't work right
	// if the umask is set to not allow world-readable permissions
	SetUmask()
}

func isMachineDebugEnabled() bool {
	return os.Getenv("MACHINE_DEBUG") != ""
}

func vmrun(args ...string) (string, string, error) {
	cmd := exec.Command(vmrunbin, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if isMachineDebugEnabled() {
		// write stdout to stderr because stdout is used for parsing sometimes
		cmd.Stdout = io.MultiWriter(os.Stderr, cmd.Stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, cmd.Stderr)
	}

	log.Debugf("executing: %v %v", vmrunbin, strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrVMRUNNotFound
		}
	}

	return stdout.String(), stderr.String(), err
}

// Make a vmdk disk image with the given size (in MB).
func vdiskmanager(dest string, size int) error {
	cmd := exec.Command(vdiskmanbin, "-c", "-t", "0", "-s", fmt.Sprintf("%dMB", size), "-a", "lsilogic", dest)
	if isMachineDebugEnabled() {
		// write stdout to stderr because stdout is used for parsing sometimes
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}

	if stdout := cmd.Run(); stdout != nil {
		if ee, ok := stdout.(*exec.Error); ok && ee == exec.ErrNotFound {
			return ErrVMRUNNotFound
		}
	}
	return nil
}
