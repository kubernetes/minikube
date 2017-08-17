package bootstrapper

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/sshutil"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"golang.org/x/crypto/ssh"
)

type CommandRunner interface {
	Run(cmd string) error
	CombinedOutput(cmd string) (string, error)
	Shell(cmd string) error

	TransferFile(assets.CopyableFile) error
	DeleteFile(assets.CopyableFile) error
}

func getDeleteFileCommand(f assets.CopyableFile) string {
	return fmt.Sprintf("sudo rm %s", filepath.Join(f.GetTargetDir(), f.GetTargetName()))
}

type ExecRunner struct{}
type SSHRunner struct {
	c *ssh.Client
}

func NewSSHRunner(c *ssh.Client) *SSHRunner {
	return &SSHRunner{c}
}

func (*ExecRunner) Run(cmd string) error {
	glog.Infoln("Run:", cmd)
	c := exec.Command("/bin/bash", "-c", cmd)
	if err := c.Run(); err != nil {
		return errors.Wrapf(err, "running command: %s", cmd)
	}
	return nil
}

func (*ExecRunner) CombinedOutput(cmd string) (string, error) {
	glog.Infoln("Run with output:", cmd)
	c := exec.Command("/bin/bash", "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "running command: %s\n output: %s", cmd, out)
	}
	return string(out), nil
}

// Make this some sort of pipe
func (e *ExecRunner) Shell(cmd string) error {
	return e.Run(cmd)
}

func (*ExecRunner) TransferFile(f assets.CopyableFile) error {
	if err := os.MkdirAll(f.GetTargetDir(), os.ModePerm); err != nil {
		return errors.Wrapf(err, "error making dirs for %s", f.GetTargetDir())
	}
	targetPath := filepath.Join(f.GetTargetDir(), f.GetTargetName())
	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Remove(targetPath); err != nil {
			return errors.Wrapf(err, "error removing file %s", targetPath)
		}

	}
	target, err := os.Create(targetPath)
	if err != nil {
		return errors.Wrapf(err, "error creating file at %s", targetPath)
	}
	perms, err := strconv.Atoi(f.GetPermissions())
	if err != nil {
		return errors.Wrapf(err, "error converting permissions %s to integer", perms)
	}
	if err := target.Chmod(os.FileMode(perms)); err != nil {
		return errors.Wrapf(err, "error changing file permissions for %s", targetPath)
	}

	if _, err = io.Copy(target, f); err != nil {
		return errors.Wrapf(err, `error copying file %s to target location:
do you have the correct permissions?  The none driver requires sudo for the "start" command`,
			targetPath)
	}
	return target.Close()
}

func (e *ExecRunner) DeleteFile(f assets.CopyableFile) error {
	cmd := getDeleteFileCommand(f)
	return e.Run(cmd)
}

func (s *SSHRunner) DeleteFile(f assets.CopyableFile) error {
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	cmd := getDeleteFileCommand(f)
	return sess.Run(cmd)
}

func (s *SSHRunner) Run(cmd string) error {
	glog.Infoln("Run:", cmd)
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	return sess.Run(cmd)
}

func (s *SSHRunner) CombinedOutput(cmd string) (string, error) {
	glog.Infoln("Run with output:", cmd)
	sess, err := s.c.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	out, err := sess.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "running command: %s\n output: %s", cmd, out)
	}
	return string(out), nil
}

func (s *SSHRunner) Shell(cmd string) error {
	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "getting ssh session")
	}
	return sshutil.GetShell(sess, cmd)
}

func (s *SSHRunner) TransferFile(f assets.CopyableFile) error {
	deleteCmd := fmt.Sprintf("sudo rm -f %s", filepath.Join(f.GetTargetDir(), f.GetTargetName()))
	mkdirCmd := fmt.Sprintf("sudo mkdir -p %s", f.GetTargetDir())
	for _, cmd := range []string{deleteCmd, mkdirCmd} {
		if err := s.Run(cmd); err != nil {
			return errors.Wrapf(err, "Error running command: %s", cmd)
		}
	}

	sess, err := s.c.NewSession()
	if err != nil {
		return errors.Wrap(err, "Error creating new session via ssh client")
	}

	w, err := sess.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "Error accessing StdinPipe via ssh session")
	}
	// The scpcmd below *should not* return until all data is copied and the
	// StdinPipe is closed. But let's use a WaitGroup to make it expicit.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer w.Close()
		header := fmt.Sprintf("C%s %d %s\n", f.GetPermissions(), f.GetLength(), f.GetTargetName())
		fmt.Fprint(w, header)
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()

	scpcmd := fmt.Sprintf("sudo scp -t %s", f.GetTargetDir())
	if err := sess.Run(scpcmd); err != nil {
		return errors.Wrapf(err, "Error running scp command: %s", scpcmd)
	}
	wg.Wait()

	return nil
}
