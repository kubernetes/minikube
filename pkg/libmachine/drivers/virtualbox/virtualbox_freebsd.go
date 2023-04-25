package virtualbox

import "path/filepath"

func detectVBoxManageCmd() string {
	return detectVBoxManageCmdInPath()
}

func getShareDriveAndName() (string, string) {
	path, err := filepath.EvalSymlinks("/home")
	if err != nil {
		path = "/home"
	}

	return "hosthome", path
}

func isHyperVInstalled() bool {
	return false
}
