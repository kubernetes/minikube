package virtualbox

func detectVBoxManageCmd() string {
	return detectVBoxManageCmdInPath()
}

func getShareDriveAndName() (string, string) {
	return "Users", "/Users"
}

func isHyperVInstalled() bool {
	return false
}
