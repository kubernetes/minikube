/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

// RunResult holds the results of a Runner
type RunResult struct {
	Stdout   bytes.Buffer
	Stderr   bytes.Buffer
	ExitCode int
	Args     []string // the args that was passed to Runner
}

// Command returns a human readable command string that does not induce eye fatigue
func (rr RunResult) Command() string {
	var sb strings.Builder
	sb.WriteString(rr.Args[0])
	for _, a := range rr.Args[1:] {
		if strings.Contains(a, " ") {
			sb.WriteString(fmt.Sprintf(` "%s"`, a))
			continue
		}
		sb.WriteString(fmt.Sprintf(" %s", a))
	}
	return sb.String()
}

// Output returns human-readable output for an execution result
func (rr RunResult) Output() string {
	var sb strings.Builder
	if rr.Stdout.Len() > 0 {
		sb.WriteString(fmt.Sprintf("-- stdout --\n%s\n-- /stdout --", rr.Stdout.Bytes()))
	}
	if rr.Stderr.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n** stderr ** \n%s\n** /stderr **", rr.Stderr.Bytes()))
	}
	return sb.String()
}

// PrefixCmd adds any needed prefix (such as sudo) to the command
func PrefixCmd(cmd *exec.Cmd) *exec.Cmd {
	if cmd.Args[0] == Podman && runtime.GOOS == "linux" { // want sudo when not running podman-remote
		cmdWithSudo := exec.Command("sudo", append([]string{"-n"}, cmd.Args...)...)
		cmdWithSudo.Env = cmd.Env
		cmdWithSudo.Dir = cmd.Dir
		cmdWithSudo.Stdin = cmd.Stdin
		cmdWithSudo.Stdout = cmd.Stdout
		cmdWithSudo.Stderr = cmd.Stderr
		cmd = cmdWithSudo
	}
	return cmd
}

// runCmd runs a command exec.Command against docker daemon or podman
func runCmd(cmd *exec.Cmd, warnSlow ...bool) (*RunResult, error) {
	cmd = PrefixCmd(cmd)

	warn := false
	if len(warnSlow) > 0 {
		warn = warnSlow[0]
	}

	killTime := 19 * time.Second // this will be applied only if warnSlow is true
	warnTime := 2 * time.Second

	if cmd.Args[1] == "volume" || cmd.Args[1] == "ps" { // volume and ps requires more time than inspect
		killTime = 30 * time.Second
		warnTime = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), killTime)
	defer cancel()

	if warn { // convert exec.Command to with context
		cmdWithCtx := exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
		cmdWithCtx.Stdout = cmd.Stdout // copying the original command
		cmdWithCtx.Stderr = cmd.Stderr
		cmd = cmdWithCtx
	}

	rr := &RunResult{Args: cmd.Args}
	klog.Infof("Run: %v", rr.Command())

	var outb, errb io.Writer
	if cmd.Stdout == nil {
		var so bytes.Buffer
		outb = io.MultiWriter(&so, &rr.Stdout)
	} else {
		outb = io.MultiWriter(cmd.Stdout, &rr.Stdout)
	}

	if cmd.Stderr == nil {
		var se bytes.Buffer
		errb = io.MultiWriter(&se, &rr.Stderr)
	} else {
		errb = io.MultiWriter(cmd.Stderr, &rr.Stderr)
	}

	cmd.Stdout = outb
	cmd.Stderr = errb

	start := time.Now()
	klog.Infof("Running %v", cmd.Args)
	err := cmd.Run()
	elapsed := time.Since(start)
	if warn {
		if elapsed > warnTime {
			out.WarningT(`Executing "{{.command}}" took an unusually long time: {{.duration}}`, out.V{"command": rr.Command(), "duration": elapsed})
			// Don't show any restarting hint, when running podman locally (on linux, with sudo). Only when having a service.
			if cmd.Args[0] != "sudo" {
				out.ErrT(style.Tip, `Restarting the {{.name}} service may improve performance.`, out.V{"name": cmd.Args[0]})
			}
		}

		if ctx.Err() == context.DeadlineExceeded {
			return rr, context.DeadlineExceeded
		}
	}

	if ex, ok := err.(*exec.ExitError); ok {
		klog.Warningf("%s returned with exit code %d", rr.Command(), ex.ExitCode())
		rr.ExitCode = ex.ExitCode()
	}

	// Decrease log spam
	if elapsed > (1 * time.Second) {
		klog.Infof("Completed: %s: (%s)", rr.Command(), elapsed)
	}
	if err == nil {
		return rr, nil
	}

	return rr, fmt.Errorf("%s: %v\nstdout:\n%s\nstderr:\n%s", rr.Command(), err, rr.Stdout.String(), rr.Stderr.String())
}
