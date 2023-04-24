package runner

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"k8s.io/minikube/pkg/minikube/assets"
)

// NOTE:
// this come from minikube/pkg/minikube/command
// I think it is better suited here....
// Runner represents an interface to run commands.
type Runner interface {
	// RunCmd runs a cmd of exec.Cmd type. allowing user to set cmd.Stdin, cmd.Stdout,...
	// not all implementors are guaranteed to handle all the properties of cmd.
	RunCmd(cmd *exec.Cmd) (*RunResult, error)

	// StartCmd starts a cmd of exec.Cmd type.
	// This func in non-blocking, use WaitCmd to block until complete.
	// Not all implementors are guaranteed to handle all the properties of cmd.
	StartCmd(cmd *exec.Cmd) (*StartedCmd, error)

	// WaitCmd will prevent further execution until the started command has completed.
	WaitCmd(startedCmd *StartedCmd) (*RunResult, error)

	// Copy is a convenience method that runs a command to copy a file
	Copy(assets.CopyableFile) error

	// CopyFrom is a convenience method that runs a command to copy a file back
	CopyFrom(assets.CopyableFile) error

	// Remove is a convenience method that runs a command to remove a file
	RemoveFile(assets.CopyableFile) error

	// ReadableFile open a remote file for reading
	ReadableFile(sourcePath string) (assets.ReadableFile, error)
}

// RunResult holds the results of a Runner
type RunResult struct {
	Stdout   bytes.Buffer
	Stderr   bytes.Buffer
	ExitCode int
	Args     []string // the args that was passed to Runner
}

// StartedCmd holds the contents of a started command
type StartedCmd struct {
	cmd *exec.Cmd
	rr  *RunResult
	wg  *sync.WaitGroup
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
