package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine"
)

func cmdURL(c CommandLine, api libmachine.API) error {
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

	url, err := host.URL()
	if err != nil {
		return err
	}

	fmt.Println(url)

	return nil
}
