package crashreport

import "os/exec"

func localOSVersion() string {
	command := exec.Command("uname", "-r")
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return string(output)
}
