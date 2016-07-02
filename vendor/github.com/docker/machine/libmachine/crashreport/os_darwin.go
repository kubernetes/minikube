package crashreport

import "os/exec"

func localOSVersion() string {
	command := exec.Command("bash", "-c", `sw_vers | grep ProductVersion | cut -d$'\t' -f2`)
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return string(output)
}
