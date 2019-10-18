package driver

import "os/exec"

// supportedDrivers is a list of supported drivers on Darwin.
var supportedDrivers = []string{
	VirtualBox,
	Parallels,
	VMwareFusion,
	HyperKit,
	VMware,
}

func VBoxManagePath() string {
	cmd := "VBoxManage"
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}
