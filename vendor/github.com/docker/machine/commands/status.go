package commands

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

func cmdStatus(c CommandLine, api libmachine.API) error {
	if len(c.Args()) > 1 {
		return ErrExpectedOneMachine
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
		log.Errorf("error getting state for host %s: %s", host.Name, err)
	}

	log.Info(currentState)

	return nil
}
