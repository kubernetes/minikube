package provision

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/mcnutils"
)

// DockerClientVersion returns the version of the Docker client inside the machine
func DockerClientVersion(prov Provisioner) (string, error) {
	// `docker version --format {{.Client.Version}}` would be preferable, but
	// that fails if the server isn't running yet.
	//
	// output is expected to be something like
	//
	//     Docker version 1.12.1, build 7a86f89
	output, err := prov.RunCommand("docker --version")
	if err != nil {
		return "", err
	}

	words := strings.Fields(output)
	if len(words) < 3 || words[0] != "Docker" || words[1] != "version" {
		return "", fmt.Errorf("DockerClientVersion: cannot parse version string from %q", output)
	}

	return strings.TrimRight(words[2], ","), nil
}

// waitForLock waits for the package manager to be available and issues a repository update
func waitForLock(prov Provisioner) error {
	return func(prov Provisioner, cmd string) error {
		var cmdErr error
		err := mcnutils.WaitFor(func() bool {
			_, cmdErr = prov.RunCommand(cmd)
			if cmdErr != nil {
				if strings.Contains(cmdErr.Error(), "Could not get lock") {
					cmdErr = nil
					return false
				}
				return true
			}
			return true
		})
		if cmdErr != nil {
			return fmt.Errorf("Error running %q: %s", cmd, cmdErr)
		}
		if err != nil {
			return fmt.Errorf("Failed to obtain lock: %s", err)
		}
		return nil
	}(prov, "sudo apt-get update") // !?
}
