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

package deployer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

const ssh = "ssh"
const rsync = "rsync"

func executeLocalCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	klog.Infof("Executing: %v", cmd.Args)
	return cmd.Run()
}

func executeSSHCommand(ctx context.Context, user string, addr string, sshArguments []string, args ...string) error {
	allArgs := []string{addr, "-o", "StrictHostKeyChecking=no",
		"-o", "User=" + user, "-o", "UserKnownHostsFile=/dev/null"}
	allArgs = append(allArgs, sshArguments...)
	allArgs = append(allArgs, "--")
	allArgs = append(allArgs, args...)
	cmd := exec.CommandContext(ctx, "ssh", allArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	klog.Infof("Executing: %v", cmd.Args)
	return cmd.Run()
}

func sshConnectionCheck(ctx context.Context, user string, addr string, sshArguments []string) error {
	var err error
	for i := range 10 {
		//  cmd cannot be reused after its failure
		err = executeSSHCommand(ctx, user, addr, sshArguments, "uname", "-a")
		if err == nil {
			return nil
		}
		klog.Infof("[%d/10]ssh command failed with error: %v", i, err)
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("failed to connect to vm: %v", err)
}

func executeRsyncSSHCommand(ctx context.Context, sshArguments []string, src string, dst string, rsyncArgs []string) error {
	sshArgs := []string{ssh, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}
	sshArgs = append(sshArgs, sshArguments...)

	allArgs := []string{"-e", strings.Join(sshArgs, " "), "-avz"}
	allArgs = append(allArgs, rsyncArgs...)
	allArgs = append(allArgs, src, dst)
	cmd := exec.CommandContext(ctx, rsync, allArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	klog.Infof("Executing: %v", cmd.Args)
	return cmd.Run()
}

func executeScpCommand(ctx context.Context, user string, addr string, sshArguments []string, src string, dst string) error {
	allArgs := []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}
	allArgs = append(allArgs, sshArguments...)

	allArgs = append(allArgs, fmt.Sprintf("%s@%s:%s", user, addr, src), dst)
	cmd := exec.CommandContext(ctx, "scp", allArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	klog.Infof("Executing: %v", cmd.Args)
	return cmd.Run()
}
