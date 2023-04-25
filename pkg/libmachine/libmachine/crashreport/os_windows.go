package crashreport

import (
	"os/exec"
	"strings"
)

func localOSVersion() string {
	command := exec.Command("ver")
	output, err := command.Output()
	if err == nil {
		return parseVerOutput(string(output))
	}

	command = exec.Command("systeminfo")
	output, err = command.Output()
	if err == nil {
		return parseSystemInfoOutput(string(output))
	}

	return ""
}

func parseSystemInfoOutput(output string) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "OS Version:") {
			return strings.TrimSpace(line[len("OS Version:"):])
		}
	}

	// If we couldn't find the version, maybe the output is not in English
	// Let's parse the fourth line since it seems to be the one always used
	// for the version.
	if len(lines) >= 4 {
		parts := strings.Split(lines[3], ":")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}

	return ""
}

func parseVerOutput(output string) string {
	return strings.TrimSpace(output)
}
