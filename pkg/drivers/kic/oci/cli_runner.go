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
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

var warnLock sync.Mutex
var alreadyWarnedCmds = make(map[string]bool)

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

// IsRootlessForced returns whether rootless mode is explicitly required.
func IsRootlessForced() bool {
	s := os.Getenv(constants.MinikubeRootlessEnv)
	if s == "" {
		return false
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		klog.ErrorS(err, "failed to parse", "env", constants.MinikubeRootlessEnv, "value", s)
		return false
	}
	return v
}

type prefixCmdOptions struct {
	sudoFlags []string
}

type PrefixCmdOption func(*prefixCmdOptions)

func WithSudoFlags(ss ...string) PrefixCmdOption {
	return func(o *prefixCmdOptions) {
		o.sudoFlags = ss
	}
}

// PrefixCmd adds any needed prefix (such as sudo) to the command
func PrefixCmd(cmd *exec.Cmd, opt ...PrefixCmdOption) *exec.Cmd {
	var o prefixCmdOptions
	for _, f := range opt {
		f(&o)
	}
	if cmd.Args[0] == Podman && runtime.GOOS == "linux" && !IsRootlessForced() { // want sudo when not running podman-remote
		cmdWithSudo := exec.Command("sudo", append(append([]string{"-n"}, o.sudoFlags...), cmd.Args...)...)
		cmdWithSudo.Env = cmd.Env
		cmdWithSudo.Dir = cmd.Dir
		cmdWithSudo.Stdin = cmd.Stdin
		cmdWithSudo.Stdout = cmd.Stdout
		cmdWithSudo.Stderr = cmd.Stderr
		cmd = cmdWithSudo
	}
	return cmd
}

func suppressDockerMessage() bool {
	envKey := "MINIKUBE_SUPPRESS_DOCKER_PERFORMANCE"
	env := os.Getenv(envKey)
	if env == "" {
		return false
	}
	suppress, err := strconv.ParseBool(env)
	if err != nil {
		msg := fmt.Sprintf("failed to parse bool from the %s env, defaulting to 'false'; received: %s: %v", envKey, env, err)
		klog.Warning(msg)
		out.Styled(style.Warning, msg)
		return false
	}
	return suppress
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

	// Increase timeout for operations that typically take longer
	if len(cmd.Args) > 1 {
		switch cmd.Args[1] {
		case "volume", "ps":
			// volume and ps requires more time than inspect
			killTime = 30 * time.Second
			warnTime = 3 * time.Second
		case "rm", "remove":
			// Container/volume removal can take longer, especially if containers are stuck
			killTime = 60 * time.Second // CHANGE: Increase from 45 to 60 seconds
			warnTime = 5 * time.Second
		case "stop", "kill":
			// Stop/kill operations can take time if containers are unresponsive
			killTime = 30 * time.Second
			warnTime = 3 * time.Second
		case "network":
			// Network operations can be slow during cleanup
			killTime = 30 * time.Second
			warnTime = 3 * time.Second
		case "prune":
			// Prune operations can take a very long time
			killTime = 120 * time.Second
			warnTime = 10 * time.Second
		}
	}

	rr := &RunResult{Args: cmd.Args}
	klog.Infof("Run: %v", rr.Command())

	// Determine the actual OCI binary name for user-facing messages
	actualBinaryName := getActualBinaryName(cmd.Args)

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
	var err error

	if warn {
		// Use context with timeout for proper cancellation
		ctx, cancel := context.WithTimeout(context.Background(), killTime)
		defer cancel()

		// Create command with context
		cmdWithCtx := exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
		cmdWithCtx.Stdout = cmd.Stdout
		cmdWithCtx.Stderr = cmd.Stderr
		cmdWithCtx.Env = cmd.Env
		cmdWithCtx.Dir = cmd.Dir

		// Channel to track completion
		done := make(chan error, 1)
		var warnTimer *time.Timer

		// Start the command in a goroutine
		go func() {
			done <- cmdWithCtx.Run()
		}()

		// Set up warning timer
		if !out.JSON && !suppressDockerMessage() {
			warnTimer = time.AfterFunc(warnTime, func() {
				warnLock.Lock()
				_, ok := alreadyWarnedCmds[rr.Command()]
				if !ok {
					alreadyWarnedCmds[rr.Command()] = true
					warnLock.Unlock()

					// Show operation-specific warning messages
					if len(cmd.Args) > 1 {
						switch cmd.Args[1] {
						case "ps":
							out.WarningT(`"{{.command}}" is taking longer than expected. This may indicate {{.name}} is hanging.`, out.V{"command": rr.Command(), "name": actualBinaryName})
						case "rm", "remove":
							out.WarningT(`"{{.command}}" is taking longer than expected. Container may be stuck - please be patient.`, out.V{"command": rr.Command()})
						case "volume":
							out.WarningT(`"{{.command}}" is taking longer than expected. Volume operations can be slow.`, out.V{"command": rr.Command()})
						case "network":
							out.WarningT(`"{{.command}}" is taking longer than expected. Network cleanup can be slow.`, out.V{"command": rr.Command()})
						case "prune":
							out.WarningT(`"{{.command}}" is taking longer than expected. Prune operations can take several minutes.`, out.V{"command": rr.Command()})
						default:
							out.WarningT(`"{{.command}}" is taking an unusually long time to respond, please be patient.`, out.V{"command": rr.Command()})
						}
					} else {
						out.WarningT(`"{{.command}}" is taking an unusually long time to respond, please be patient.`, out.V{"command": rr.Command()})
					}

					// Show restart hint using the actual binary name
					out.ErrT(style.Tip, `If this continues to hang, consider restarting the {{.name}} service.`, out.V{"name": actualBinaryName})
				} else {
					warnLock.Unlock()
				}
			})
		}

		// Wait for completion or timeout
		select {
		case err = <-done:
			// Command completed normally
			if warnTimer != nil {
				warnTimer.Stop()
			}
		case <-ctx.Done():
			// Command timed out
			if warnTimer != nil {
				warnTimer.Stop()
			}

			// Kill the process if it's still running
			if cmdWithCtx.Process != nil {
				klog.Warningf("Killing slow %s process after %v timeout", rr.Command(), killTime)
				cmdWithCtx.Process.Kill()
			}

			out.WarningT(`"{{.command}}" took too long to respond (>{{.duration}}) and was terminated.`, out.V{"command": rr.Command(), "duration": killTime})
			out.ErrT(style.Tip, `Consider restarting the {{.name}} service if this problem persists.`, out.V{"name": actualBinaryName})

			return rr, fmt.Errorf("command timed out after %v: %s", killTime, rr.Command())
		}
	} else {
		// Run without timeout for non-critical operations
		err = cmd.Run()
	}

	elapsed := time.Since(start)

	// Log completion information
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

// Helper function to extract the actual OCI binary name from a command
func getActualBinaryName(cmdArgs []string) string {
	if len(cmdArgs) == 0 {
		return ""
	}

	// If not using sudo, return the first argument
	// Note: This checks for "sudo" specifically because PrefixCmd()
	// explicitly adds "sudo" as a string literal when needed
	if cmdArgs[0] != "sudo" {
		return cmdArgs[0]
	}

	// Parse sudo command to find the actual binary
	for i := 1; i < len(cmdArgs); i++ {
		arg := cmdArgs[i]
		// Skip sudo flags (those starting with -)
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}

	return cmdArgs[0] // fallback to sudo if we can't parse
}
