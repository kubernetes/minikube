package virtualbox

import (
	"bytes"
	"io/ioutil"

	"github.com/docker/machine/libmachine/log"
)

// IsVTXDisabled checks if VT-X is disabled in the BIOS. If it is, the vm will fail to start.
// If we can't be sure it is disabled, we carry on and will check the vm logs after it's started.
// We want to check that either vmx or svm flags are present in /proc/cpuinfo.
func (d *Driver) IsVTXDisabled() bool {
	cpuinfo, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		log.Debugf("Couldn't check that VT-X/AMD-v is enabled. Will check that the vm is properly created: %v", err)
		return false
	}
	return isVTXDisabled(cpuinfo)
}

func isVTXDisabled(cpuinfo []byte) bool {
	features := [2][]byte{
		{'v', 'm', 'x'},
		{'s', 'v', 'm'},
	}
	for _, v := range features {
		if bytes.Contains(cpuinfo, v) {
			return false
		}
	}
	return true
}

func detectVBoxManageCmd() string {
	return detectVBoxManageCmdInPath()
}

func getShareDriveAndName() (string, string) {
	return "hosthome", "/home"
}

func isHyperVInstalled() bool {
	return false
}
