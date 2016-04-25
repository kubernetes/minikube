package virtualbox

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"strconv"

	"time"

	"github.com/docker/machine/libmachine/log"
)

const (
	retryCountOnObjectNotReadyError = 5
	objectNotReady                  = "error: The object is not ready"
	retryDelay                      = 100 * time.Millisecond
)

var (
	reColonLine       = regexp.MustCompile(`(.+):\s+(.*)`)
	reEqualLine       = regexp.MustCompile(`(.+)=(.*)`)
	reEqualQuoteLine  = regexp.MustCompile(`"(.+)"="(.*)"`)
	reMachineNotFound = regexp.MustCompile(`Could not find a registered machine named '(.+)'`)

	ErrMachineNotExist = errors.New("machine does not exist")
	ErrVBMNotFound     = errors.New("VBoxManage not found. Make sure VirtualBox is installed and VBoxManage is in the path")

	vboxManageCmd = detectVBoxManageCmd()
)

// VBoxManager defines the interface to communicate to VirtualBox.
type VBoxManager interface {
	vbm(args ...string) error

	vbmOut(args ...string) (string, error)

	vbmOutErr(args ...string) (string, string, error)
}

// VBoxCmdManager communicates with VirtualBox through the commandline using `VBoxManage`.
type VBoxCmdManager struct {
	runCmd func(cmd *exec.Cmd) error
}

// NewVBoxManager creates a VBoxManager instance.
func NewVBoxManager() *VBoxCmdManager {
	return &VBoxCmdManager{
		runCmd: func(cmd *exec.Cmd) error { return cmd.Run() },
	}
}

func (v *VBoxCmdManager) vbm(args ...string) error {
	_, _, err := v.vbmOutErr(args...)
	return err
}

func (v *VBoxCmdManager) vbmOut(args ...string) (string, error) {
	stdout, _, err := v.vbmOutErr(args...)
	return stdout, err
}

func (v *VBoxCmdManager) vbmOutErr(args ...string) (string, string, error) {
	return v.vbmOutErrRetry(retryCountOnObjectNotReadyError, args...)
}

func (v *VBoxCmdManager) vbmOutErrRetry(retry int, args ...string) (string, string, error) {
	cmd := exec.Command(vboxManageCmd, args...)
	log.Debugf("COMMAND: %v %v", vboxManageCmd, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := v.runCmd(cmd)
	stderrStr := stderr.String()
	if len(args) > 0 {
		log.Debugf("STDOUT:\n{\n%v}", stdout.String())
		log.Debugf("STDERR:\n{\n%v}", stderrStr)
	}

	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee.Err == exec.ErrNotFound {
			err = ErrVBMNotFound
		}
	}

	// Sometimes, we just need to retry...
	if retry > 1 {
		if strings.Contains(stderrStr, objectNotReady) {
			time.Sleep(retryDelay)
			return v.vbmOutErrRetry(retry-1, args...)
		}
	}

	if err == nil || strings.HasPrefix(err.Error(), "exit status ") {
		// VBoxManage will sometimes not set the return code, but has a fatal error
		// such as VBoxManage.exe: error: VT-x is not available. (VERR_VMX_NO_VMX)
		if strings.Contains(stderrStr, "error:") {
			err = fmt.Errorf("%v %v failed:\n%v", vboxManageCmd, strings.Join(args, " "), stderrStr)
		}
	}

	return stdout.String(), stderrStr, err
}

func checkVBoxManageVersion(version string) error {
	major, minor, err := parseVersion(version)
	if (err != nil) || (major < 4) || (major == 4 && minor <= 2) {
		return fmt.Errorf("We support Virtualbox starting with version 5. Your VirtualBox install is %q. Please upgrade at https://www.virtualbox.org", version)
	}

	if major < 5 {
		log.Warnf("You are using version %s of VirtualBox. If you encounter issues, you might want to upgrade to version 5 at https://www.virtualbox.org", version)
	}

	return nil
}

func parseVersion(version string) (int, int, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("Invalid version: %q", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid version: %q", version)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid version: %q", version)
	}

	return major, minor, err
}

func parseKeyValues(stdOut string, regexp *regexp.Regexp, callback func(key, val string) error) error {
	r := strings.NewReader(stdOut)
	s := bufio.NewScanner(r)

	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}

		res := regexp.FindStringSubmatch(line)
		if res == nil {
			continue
		}

		key, val := res[1], res[2]
		if err := callback(key, val); err != nil {
			return err
		}
	}

	return s.Err()
}
