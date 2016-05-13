package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
)

type errStateInvalidForSSH struct {
	HostName string
}

func (e errStateInvalidForSSH) Error() string {
	return fmt.Sprintf("Error: Cannot run SSH command: Host %q is not running", e.HostName)
}

func cmdSSH(c CommandLine, api libmachine.API) error {
	// Check for help flag -- Needed due to SkipFlagParsing
	firstArg := c.Args().First()
	if firstArg == "-help" || firstArg == "--help" || firstArg == "-h" {
		c.ShowHelp()
		return nil
	}

	target, err := targetHost(c, api)
	if err != nil {
		return err
	}

	host, err := api.Load(target)
	if err != nil {
		return err
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return err
	}

	if currentState != state.Running {
		return errStateInvalidForSSH{host.Name}
	}

	client, err := host.CreateSSHClient()
	if err != nil {
		return err
	}

	return client.Shell(c.Args().Tail()...)
}
