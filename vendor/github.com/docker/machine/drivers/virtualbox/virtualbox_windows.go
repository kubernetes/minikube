package virtualbox

import (
	"strings"

	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/machine/libmachine/log"
	"golang.org/x/sys/windows/registry"
)

// IsVTXDisabled checks if VT-X is disabled in the BIOS. If it is, the vm will fail to start.
// If we can't be sure it is disabled, we carry on and will check the vm logs after it's started.
func (d *Driver) IsVTXDisabled() bool {
	errmsg := "Couldn't check that VT-X/AMD-v is enabled. Will check that the vm is properly created: %v"
	output, err := cmdOutput("wmic", "cpu", "get", "VirtualizationFirmwareEnabled")
	if err != nil {
		log.Debugf(errmsg, err)
		return false
	}

	disabled := strings.Contains(output, "FALSE")
	return disabled
}

// cmdOutput runs a shell command and returns its output.
func cmdOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	log.Debugf("COMMAND: %v %v", name, strings.Join(args, " "))

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	log.Debugf("STDOUT:\n{\n%v}", string(stdout))

	return string(stdout), nil
}

func detectVBoxManageCmd() string {
	cmd := "VBoxManage"
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

	// Look in default installation path for VirtualBox version > 5
	if path, err := exec.LookPath(filepath.Join("C:\\Program Files\\Oracle\\VirtualBox", cmd)); err == nil {
		return path
	}

	// Look in windows registry
	if p, err := findVBoxInstallDirInRegistry(); err == nil {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path
		}
	}

	return detectVBoxManageCmdInPath() //fallback to path
}

func findVBoxInstallDirInRegistry() (string, error) {
	registryKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Oracle\VirtualBox`, registry.QUERY_VALUE)
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find VirtualBox registry entries, is VirtualBox really installed properly? %s", err)
		log.Debugf(errorMessage)
		return "", fmt.Errorf(errorMessage)
	}

	defer registryKey.Close()

	installDir, _, err := registryKey.GetStringValue("InstallDir")
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find InstallDir registry key within VirtualBox registries entries, is VirtualBox really installed properly? %s", err)
		log.Debugf(errorMessage)
		return "", fmt.Errorf(errorMessage)
	}

	return installDir, nil
}

func getShareDriveAndName() (string, string) {
	return "c/Users", "\\\\?\\c:\\Users"
}

func isHyperVInstalled() bool {
	// check if hyper-v is installed
	_, err := exec.LookPath("vmms.exe")
	if err != nil {
		errmsg := "Hyper-V is not installed."
		log.Debugf(errmsg, err)
		return false
	}

	// check to see if a hypervisor is present. if hyper-v is installed and enabled,
	// display an error explaining the incompatibility between virtualbox and hyper-v.
	output, err := cmdOutput("wmic", "computersystem", "get", "hypervisorpresent")

	if err != nil {
		errmsg := "Could not check to see if Hyper-V is running."
		log.Debugf(errmsg, err)
		return false
	}

	enabled := strings.Contains(output, "TRUE")
	return enabled

}
