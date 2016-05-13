package commands

import (
	"fmt"

	"strings"

	"errors"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

func cmdRm(c CommandLine, api libmachine.API) error {
	if len(c.Args()) == 0 {
		c.ShowHelp()
		return ErrNoMachineSpecified
	}

	log.Info(fmt.Sprintf("About to remove %s", strings.Join(c.Args(), ", ")))

	force := c.Bool("force")
	confirm := c.Bool("y")
	var errorOccured []string

	if !userConfirm(confirm, force) {
		return nil
	}

	for _, hostName := range c.Args() {
		err := removeRemoteMachine(hostName, api)
		if err != nil {
			errorOccured = collectError(fmt.Sprintf("Error removing host %q: %s", hostName, err), force, errorOccured)
		}

		if err == nil || force {
			removeErr := removeLocalMachine(hostName, api)
			if removeErr != nil {
				errorOccured = collectError(fmt.Sprintf("Can't remove \"%s\"", hostName), force, errorOccured)
			} else {
				log.Infof("Successfully removed %s", hostName)
			}
		}
	}

	if len(errorOccured) > 0 && !force {
		return errors.New(strings.Join(errorOccured, "\n"))
	}

	return nil
}

func userConfirm(confirm bool, force bool) bool {
	if confirm || force {
		return true
	}

	sure, err := confirmInput(fmt.Sprintf("Are you sure?"))
	if err != nil {
		return false
	}

	return sure
}

func removeRemoteMachine(hostName string, api libmachine.API) error {
	currentHost, loaderr := api.Load(hostName)
	if loaderr != nil {
		return loaderr
	}

	return currentHost.Driver.Remove()
}

func removeLocalMachine(hostName string, api libmachine.API) error {
	exist, _ := api.Exists(hostName)
	if !exist {
		return errors.New(hostName + " does not exist.")
	}
	return api.Remove(hostName)
}

func collectError(message string, force bool, errorOccured []string) []string {
	if force {
		log.Error(message)
	}
	return append(errorOccured, message)
}
