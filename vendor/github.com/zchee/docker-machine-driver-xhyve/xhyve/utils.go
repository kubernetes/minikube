package xhyve

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/machine/libmachine/log"
)

var (
	ErrDdNotFound      = errors.New("xhyve not found")
	ErrUuidgenNotFound = errors.New("uuidgen not found")
	ErrHdiutilNotFound = errors.New("hdiutil not found")
	ErrVBMNotFound     = errors.New("VBoxManage not found")
	vboxManageCmd      = setVBoxManageCmd()
)

func hdiutil(args ...string) error {
	cmd := exec.Command("hdiutil", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Debugf("executing: %v %v", cmd, strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.Chmod(dst, fi.Mode()); err != nil {
		return err
	}

	return nil
}

// detect the VBoxManage cmd's path if needed
func setVBoxManageCmd() string {
	cmd := "VBoxManage"
	path, err := exec.LookPath(cmd)
	if err != nil {
		return ""
	} else if err == nil {
		return path
	}
	if runtime.GOOS == "windows" {
		if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" {
			if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
				return path
			}
		}
		if p := os.Getenv("VBOX_MSI_INSTALL_PATH"); p != "" {
			if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
				return path
			}
		}
		// look at HKEY_LOCAL_MACHINE\SOFTWARE\Oracle\VirtualBox\InstallDir
		p := "C:\\Program Files\\Oracle\\VirtualBox"
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}
	return cmd
}

func vbm(args ...string) error {
	_, _, err := vbmOutErr(args...)
	return err
}

func vbmOut(args ...string) (string, error) {
	stdout, _, err := vbmOutErr(args...)
	return stdout, err
}

func vbmOutErr(args ...string) (string, string, error) {
	cmd := exec.Command(vboxManageCmd, args...)
	log.Debugf("executing: %v %v", vboxManageCmd, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stderrStr := stderr.String()
	log.Debugf("STDOUT: %v", stdout.String())
	log.Debugf("STDERR: %v", stderrStr)
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrVBMNotFound
		}
	} else {
		// VBoxManage will sometimes not set the return code, but has a fatal error
		// such as VBoxManage.exe: error: VT-x is not available. (VERR_VMX_NO_VMX)
		if strings.Contains(stderrStr, "error:") {
			err = fmt.Errorf("%v %v failed: %v", vboxManageCmd, strings.Join(args, " "), stderrStr)
		}
	}
	return stdout.String(), stderrStr, err
}

func vboxVersionDetect() (string, error) {
	if vboxManageCmd == "" {
		return "", nil
	}
	ver, err := vbmOut("-v")
	if err != nil {
		return "", err
	}
	return ver, err
}

func toPtr(s string) *string {
	return &s
}
