package commands

import (
	"fmt"
	"k8s.io/minikube/pkg/libmachine/libmachine"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
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
		return fmt.Errorf("error getting state for host %s: %s", host.Name, err)
	}

	log.Info(currentState)

	return nil
}
