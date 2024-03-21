/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package sshagent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-ps"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
)

type sshAgent struct {
	authSock string
	agentPID int
}

// Start an ssh-agent process
func Start(profile string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("starting an SSH agent on Windows is not yet supported")
	}
	cc, err := config.Load(profile)
	if err != nil {
		return fmt.Errorf("failed loading config: %v", err)
	}
	running, err := isRunning(cc)
	if err != nil {
		return fmt.Errorf("failed checking if SSH agent is running: %v", err)
	}
	if running {
		klog.Info("SSH agent is already running, aborting start")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ssh-agent").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed starting ssh-agent: %s: %v", string(out), err)
	}
	parsed, err := parseOutput(string(out))
	if err != nil {
		return fmt.Errorf("failed to parse ssh-agent output: %v", err)
	}
	cc.SSHAuthSock = parsed.authSock
	cc.SSHAgentPID = parsed.agentPID
	if err := config.Write(profile, cc); err != nil {
		return fmt.Errorf("failed writing config: %v", err)
	}
	return nil
}

func parseOutput(out string) (*sshAgent, error) {
	sockSubmatches := regexp.MustCompile(`SSH_AUTH_SOCK=(.*?);`).FindStringSubmatch(out)
	if len(sockSubmatches) < 2 {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not found in output: %s", out)
	}
	pidSubmatches := regexp.MustCompile(`SSH_AGENT_PID=(.*?);`).FindStringSubmatch(out)
	if len(pidSubmatches) < 2 {
		return nil, fmt.Errorf("SSH_AGENT_PID not found in output: %s", out)
	}
	pid, err := strconv.Atoi(pidSubmatches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to convert pid to int: %v", err)
	}
	return &sshAgent{sockSubmatches[1], pid}, nil
}

func isRunning(cc *config.ClusterConfig) (bool, error) {
	if cc.SSHAgentPID == 0 {
		return false, nil
	}
	entry, err := ps.FindProcess(cc.SSHAgentPID)
	if err != nil {
		return false, fmt.Errorf("failed finding process: %v", err)
	}
	if entry == nil {
		return false, nil
	}
	return strings.Contains(entry.Executable(), "ssh-agent"), nil
}

// Stop an ssh-agent process
func Stop(profile string) error {
	cc, err := config.Load(profile)
	if err != nil {
		return err
	}
	running, err := isRunning(cc)
	if err != nil {
		return fmt.Errorf("failed checking if SSH agent is running: %v", err)
	}
	if running {
		if err := killProcess(cc.SSHAgentPID); err != nil {
			return fmt.Errorf("failed killing SSH agent process: %v", err)
		}
	}
	cc.SSHAuthSock = ""
	cc.SSHAgentPID = 0
	if err := config.Write(profile, cc); err != nil {
		return fmt.Errorf("failed writing config: %v", err)
	}
	return nil
}

func killProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed finding process: %v", err)
	}
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("failed killing process: %v", err)
	}
	return nil
}
