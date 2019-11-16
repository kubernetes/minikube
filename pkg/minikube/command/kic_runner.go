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

package command

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/golang/glog"
	kicassets "github.com/medyagh/kic/pkg/assets"
	"github.com/medyagh/kic/pkg/command"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

// kicRunner runs commands inside a container
// It implements the CommandRunner interface.
type kicRunner struct {
	nameOrID string
	ociBin   string
}

// NewKICRunner returns a kicRunner implementor of runner which runs cmds inside a container
func NewKICRunner(containerNameOrID string, oci string) command.Runner {
	return &kicRunner{
		nameOrID: containerNameOrID,
		ociBin:   oci, // docker or podman
	}
}

func (k *kicRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	args := []string{
		"exec",
		// run with privileges so we can remount etc..
		"--privileged",
	}
	if cmd.Stdin != nil {
		args = append(args,
			"-i", // interactive so we can supply input
		)
	}
	// if the command is hooked up to the processes's output we want a tty
	if isTerminal(cmd.Stderr) || isTerminal(cmd.Stdout) {
		args = append(args,
			"-t",
		)
	}
	// set env
	for _, env := range cmd.Env {
		args = append(args, "-e", env)
	}
	// specify the container and command, after this everything will be
	// args the the command in the container rather than to docker
	args = append(
		args,
		k.nameOrID, // ... against the container
	)

	args = append(
		args,
		cmd.Args...,
	)
	cmd2 := exec.Command(k.ociBin, args...)
	cmd2.Stdin = cmd.Stdin
	cmd2.Stdout = cmd.Stdout
	cmd2.Stderr = cmd.Stderr
	cmd2.Env = cmd.Env

	rr := &command.RunResult{Args: cmd.Args}

	var outb, errb io.Writer
	if cmd2.Stdout == nil {
		var so bytes.Buffer
		outb = io.MultiWriter(&so, &rr.Stdout)
	} else {
		outb = io.MultiWriter(cmd2.Stdout, &rr.Stdout)
	}

	if cmd2.Stderr == nil {
		var se bytes.Buffer
		errb = io.MultiWriter(&se, &rr.Stderr)
	} else {
		errb = io.MultiWriter(cmd2.Stderr, &rr.Stderr)
	}

	cmd2.Stdout = outb
	cmd2.Stderr = errb

	start := time.Now()

	err := cmd2.Run()
	elapsed := time.Since(start)
	if err == nil {
		// Reduce log spam
		if elapsed > (1 * time.Second) {
			glog.Infof("(kicRunner) Done: %v: (%s)", cmd2.Args, elapsed)
		}
	} else {
		if exitError, ok := err.(*exec.ExitError); ok {
			rr.ExitCode = exitError.ExitCode()
		}
		fmt.Printf("(medya dbg) (kicRunner) Non-zero exit: %v: %v (%s)\n", cmd2.Args, err, elapsed)
		fmt.Printf("(medya dbg) (kicRunner) Output:\n %q \n", rr.Output())
		err = errors.Wrapf(err, "command failed: %s", cmd2.Args)
	}
	return rr, err

}

// Copy copies a file and its permissions
func (k *kicRunner) Copy(f kicassets.LegacyCopyableFile) error {
	assetName := f.GetAssetName()
	if _, err := os.Stat(f.GetAssetName()); os.IsNotExist(err) {
		fc := make([]byte, f.GetLength()) // Read  asset file into a []byte
		if _, err := f.Read(fc); err != nil {
			return errors.Wrap(err, "can't copy non-existing file")
		} else { // we have a MemoryAsset, will write to disk before copying
			tmpFile, err := ioutil.TempFile(os.TempDir(), "tmpf-memory-asset")
			if err != nil {
				return errors.Wrap(err, "creating temporary file")
			}
			//  clean up the temp file
			defer os.Remove(tmpFile.Name())
			if _, err = tmpFile.Write(fc); err != nil {
				return errors.Wrap(err, "write to temporary file")
			}

			// Close the file
			if err := tmpFile.Close(); err != nil {
				return errors.Wrap(err, "close temporary file")
			}
			assetName = tmpFile.Name()

		}
	}
	// based of format of "docker cp containerName:destination"
	destination := fmt.Sprintf("%s:%s/%s", k.nameOrID, f.GetTargetDir(), f.GetTargetName())
	// make sure dir exists inside the container
	if _, err := k.RunCmd(exec.Command("mkdir", "-p", f.GetTargetDir())); err != nil {
		return errors.Wrapf(err, "error making dir %s", f.GetTargetDir())
	}

	if out, err := exec.Command(k.ociBin, "cp", assetName, destination).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "error copying %s into node, output: %s", f.GetAssetName(), string(out))
	}

	if _, err := k.RunCmd(exec.Command("chmod", f.GetPermissions(), f.GetTargetDir())); err != nil {
		return errors.Wrap(err, "failed to chmod file permissions")
	}
	return nil
}

// Remove removes a file
func (k *kicRunner) Remove(f kicassets.LegacyCopyableFile) error {
	return fmt.Errorf("not implemented yet for kic runner")
}

// isTerminal returns true if the writer w is a terminal
func isTerminal(w io.Writer) bool {
	if v, ok := (w).(*os.File); ok {
		return terminal.IsTerminal(int(v.Fd()))
	}
	return false
}
