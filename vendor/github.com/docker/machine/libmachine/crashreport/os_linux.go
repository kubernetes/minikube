package crashreport

import "os/exec"

func localOSVersion() string {
	command := exec.Command("bash", "-c", `cat /etc/os-release | grep 'VERSION=' | cut -d'=' -f2`)
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return string(output)
}
