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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8s.io/minikube/pkg/libmachine/log"
)

type diskType int

const (
	diskTypeGrowable     diskType = 0
	diskTypePreallocated diskType = 2
)

var (
	vmrunbin    = setVmwareCmd("vmrun")
	vdiskmanbin = setVmwareCmd("vmware-vdiskmanager")
)

var (
	ErrMachineExist     = errors.New("machine already exists")
	ErrMachineNotExist  = errors.New("machine does not exist")
	ErrVMRUNNotFound    = errors.New("vmrun.exe not found")
	ErrVDISKMANNotFound = errors.New("vmware-vdiskmanager.exe not found")
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
	return vmrunCmd(cmd)
}

func vmrunWait(timeout time.Duration, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, vmrunbin, args...)
	return vmrunCmd(cmd)
}

func vmrunCmd(cmd *exec.Cmd) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if isMachineDebugEnabled() {
		// write stdout to stderr because stdout is used for parsing sometimes
		cmd.Stdout = io.MultiWriter(os.Stderr, cmd.Stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, cmd.Stderr)
	}

	log.Debugf("executing: %v", strings.Join(cmd.Args, " "))

	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrVMRUNNotFound
		}
	}

	return stdout.String(), stderr.String(), err
}

func vdiskmanager(args ...string) error {
	cmd := exec.Command(vdiskmanbin, args...)
	if isMachineDebugEnabled() {
		// write stdout to stderr because stdout is used for parsing sometimes
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}

	if stdout := cmd.Run(); stdout != nil {
		if ee, ok := stdout.(*exec.Error); ok && ee == exec.ErrNotFound {
			return ErrVDISKMANNotFound
		}
	}
	return nil
}

// Make a vmdk disk image with the given size (in MB).
func createDisk(path string, sizeInMB int, diskType diskType) error {
	return vdiskmanager("-c", "-t", fmt.Sprintf("%d", diskType), "-s", fmt.Sprintf("%dMB", sizeInMB), "-a", "lsilogic", path)
}

func convertDisk(srcPath string, destPath string, diskType diskType) error {
	return vdiskmanager("-r", srcPath, "-t", fmt.Sprintf("%d", diskType), destPath)
}

func growDisk(path string, sizeInMB int) error {
	return vdiskmanager("-x", fmt.Sprintf("%dMB", sizeInMB), path)
}
