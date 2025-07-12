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

package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
)

// DockerExecRunner runs commands through docker exec for remote contexts
type DockerExecRunner struct {
	containerName string
}

// NewDockerExecRunner returns a new DockerExecRunner
func NewDockerExecRunner(containerName string) *DockerExecRunner {
	return &DockerExecRunner{containerName: containerName}
}

// RunCmd implements CommandRunner interface using docker exec
func (r *DockerExecRunner) RunCmd(cmd *exec.Cmd) (*RunResult, error) {
	rr := &RunResult{Args: cmd.Args}
	start := time.Now()

	// Build docker exec command
	dockerArgs := []string{"exec"}

	// Add working directory if specified
	if cmd.Dir != "" {
		dockerArgs = append(dockerArgs, "-w", cmd.Dir)
	}

	// Add environment variables
	for _, env := range cmd.Env {
		dockerArgs = append(dockerArgs, "-e", env)
	}

	// Add container name and command
	dockerArgs = append(dockerArgs, r.containerName)
	dockerArgs = append(dockerArgs, cmd.Args...)

	dockerCmd := exec.Command("docker", dockerArgs...)

	var outb, errb bytes.Buffer
	if cmd.Stdout == nil {
		dockerCmd.Stdout = io.MultiWriter(&outb, &rr.Stdout)
	} else {
		dockerCmd.Stdout = io.MultiWriter(cmd.Stdout, &rr.Stdout)
	}

	if cmd.Stderr == nil {
		dockerCmd.Stderr = io.MultiWriter(&errb, &rr.Stderr)
	} else {
		dockerCmd.Stderr = io.MultiWriter(cmd.Stderr, &rr.Stderr)
	}

	if cmd.Stdin != nil {
		dockerCmd.Stdin = cmd.Stdin
	}

	klog.Infof("Run: %v", rr.Command())
	err := dockerCmd.Run()
	elapsed := time.Since(start)

	if exitError, ok := err.(*exec.ExitError); ok {
		rr.ExitCode = exitError.ExitCode()
	}

	if elapsed > (1 * time.Second) {
		klog.Infof("Completed: %s: (%s)", rr.Command(), elapsed)
	}

	if err == nil {
		return rr, nil
	}

	return rr, fmt.Errorf("%s: %v\nstdout:\n%s\nstderr:\n%s", rr.Command(), err, rr.Stdout.String(), rr.Stderr.String())
}

// StartCmd starts a command in the background
func (r *DockerExecRunner) StartCmd(cmd *exec.Cmd) (*StartedCmd, error) {
	dockerArgs := []string{"exec", "-d"}

	if cmd.Dir != "" {
		dockerArgs = append(dockerArgs, "-w", cmd.Dir)
	}

	for _, env := range cmd.Env {
		dockerArgs = append(dockerArgs, "-e", env)
	}

	dockerArgs = append(dockerArgs, r.containerName)
	dockerArgs = append(dockerArgs, cmd.Args...)

	dockerCmd := exec.Command("docker", dockerArgs...)

	rr := &RunResult{Args: cmd.Args}
	sc := &StartedCmd{cmd: dockerCmd, rr: rr}

	klog.Infof("Start: %v", rr.Command())

	err := dockerCmd.Start()
	return sc, err
}

// WaitCmd waits for a started command to finish
func (r *DockerExecRunner) WaitCmd(sc *StartedCmd) (*RunResult, error) {
	err := sc.cmd.Wait()

	if exitError, ok := err.(*exec.ExitError); ok {
		sc.rr.ExitCode = exitError.ExitCode()
	}

	if err == nil {
		return sc.rr, nil
	}

	return sc.rr, fmt.Errorf("%s: %v", sc.rr.Command(), err)
}

// Copy copies a file to the container
func (r *DockerExecRunner) Copy(f assets.CopyableFile) error {
	dst := path.Join(f.GetTargetDir(), f.GetTargetName())
	src := f.GetSourcePath()

	// Handle memory assets by writing to temp file first
	if src == assets.MemorySource {
		klog.Infof("docker cp memory asset --> %s:%s", r.containerName, dst)
		tf, err := os.CreateTemp("", "tmpf-memory-asset")
		if err != nil {
			return errors.Wrap(err, "creating temporary file")
		}
		defer os.Remove(tf.Name())
		defer tf.Close()

		// Write content to temp file
		if _, err := io.Copy(tf, f); err != nil {
			return errors.Wrap(err, "copying memory asset to temp file")
		}
		if err := tf.Close(); err != nil {
			return errors.Wrap(err, "closing temp file")
		}

		src = tf.Name()
	}

	klog.Infof("docker cp %s --> %s:%s", src, r.containerName, dst)

	// First ensure target directory exists
	mkdirCmd := exec.Command("docker", "exec", r.containerName, "sudo", "mkdir", "-p", f.GetTargetDir())
	if err := mkdirCmd.Run(); err != nil {
		return errors.Wrapf(err, "mkdir %s", f.GetTargetDir())
	}

	// Copy file using docker cp
	cpCmd := exec.Command("docker", "cp", src, fmt.Sprintf("%s:%s", r.containerName, dst))
	if err := cpCmd.Run(); err != nil {
		return errors.Wrapf(err, "docker cp %s", src)
	}

	// Set permissions
	chmodCmd := exec.Command("docker", "exec", r.containerName, "sudo", "chmod", f.GetPermissions(), dst)
	if err := chmodCmd.Run(); err != nil {
		return errors.Wrapf(err, "chmod %s", dst)
	}

	// Set modtime if available
	mtime, err := f.GetModTime()
	if err != nil {
		klog.Infof("error getting modtime for %s: %v", dst, err)
	} else if mtime != (time.Time{}) {
		touchCmd := exec.Command("docker", "exec", r.containerName, "sudo", "touch", "-d", mtime.Format("2006-01-02 15:04:05"), dst)
		if err := touchCmd.Run(); err != nil {
			klog.Warningf("failed to set modtime: %v", err)
		}
	}

	return nil
}

// CopyFrom copies a file from the container
func (r *DockerExecRunner) CopyFrom(f assets.CopyableFile) error {
	src := path.Join(f.GetTargetDir(), f.GetTargetName())
	dst := f.GetSourcePath()

	klog.Infof("docker cp %s:%s --> %s", r.containerName, src, dst)

	cpCmd := exec.Command("docker", "cp", fmt.Sprintf("%s:%s", r.containerName, src), dst)
	return cpCmd.Run()
}

// Remove removes a file from the container
func (r *DockerExecRunner) Remove(f assets.CopyableFile) error {
	dst := path.Join(f.GetTargetDir(), f.GetTargetName())
	klog.Infof("rm: %s", dst)

	rmCmd := exec.Command("docker", "exec", r.containerName, "sudo", "rm", dst)
	return rmCmd.Run()
}

// ReadableFile returns a readable file from the container
func (r *DockerExecRunner) ReadableFile(sourcePath string) (assets.ReadableFile, error) {
	// Get file info
	statCmd := exec.Command("docker", "exec", r.containerName, "stat", "-c", "%#a %s %y", sourcePath)
	output, err := statCmd.Output()
	if err != nil {
		return nil, errors.Wrapf(err, "stat %s", sourcePath)
	}

	parts := strings.Fields(string(output))
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected stat output: %s", output)
	}

	// Create cat command
	catCmd := exec.Command("docker", "exec", r.containerName, "cat", sourcePath)
	reader, err := catCmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "stdout pipe")
	}

	if err := catCmd.Start(); err != nil {
		return nil, errors.Wrap(err, "start cat")
	}

	// Return simple reader wrapper
	return &simpleReadableFile{
		reader:      reader,
		sourcePath:  sourcePath,
		permissions: parts[0],
	}, nil
}

type simpleReadableFile struct {
	reader      io.Reader
	sourcePath  string
	permissions string
}

func (s *simpleReadableFile) GetLength() int {
	return 0 // Not implemented for simplicity
}

func (s *simpleReadableFile) GetSourcePath() string {
	return s.sourcePath
}

func (s *simpleReadableFile) GetPermissions() string {
	return s.permissions
}

func (s *simpleReadableFile) GetModTime() (time.Time, error) {
	return time.Time{}, nil
}

func (s *simpleReadableFile) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *simpleReadableFile) Seek(_ int64, _ int) (int64, error) {
	return 0, fmt.Errorf("seek not implemented")
}

func (s *simpleReadableFile) Close() error {
	if closer, ok := s.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}