package drivers

import (
	"fmt"
	"os/exec"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
)

// TODO:
// this is some slow logic... at least make it non-blocking..
// WaitForPrompt tries to run a command to the machine shell
// for 30 seconds before timing out
func WaitForPrompt(d Driver) error {
	if err := mcnutils.WaitFor(promptAvailFunc(d)); err != nil {
		return fmt.Errorf("Too many retries waiting for prompt to be available.  Last error: %s", err)
	}
	return nil
}

// promptAvailFunc tries to run an `exit 0` shell command once,
// returns false if command fails for some reason,
// returns true if commands succeeds
func promptAvailFunc(d Driver) func() bool {
	return func() bool {
		log.Debug("Getting to WaitForPrompt function...")
		if _, err := d.GetRunner().RunCmd(exec.Command("exit", "0")); err != nil {
			log.Debugf("Error running 'exit 0' : %s", err)
			return false
		}
		return true
	}
}
