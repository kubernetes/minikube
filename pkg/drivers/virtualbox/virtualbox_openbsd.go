package virtualbox

func detectVBoxManageCmd() string {
	return detectVBoxManageCmdInPath()
}

func getShareDriveAndName() (string, string) {
	return "hosthome", "/home"
}

func isHyperVInstalled() bool {
	return false
}
