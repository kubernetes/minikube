package commands

import (
	"fmt"
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/check"
	"github.com/docker/machine/libmachine/log"
)

func cmdConfig(c CommandLine, api libmachine.API) error {
	// Ensure that log messages always go to stderr when this command is
	// being run (it is intended to be run in a subshell)
	log.SetOutWriter(os.Stderr)

	target, err := targetHost(c, api)
	if err != nil {
		return err
	}

	host, err := api.Load(target)
	if err != nil {
		return err
	}

	dockerHost, authOptions, err := check.DefaultConnChecker.Check(host, c.Bool("swarm"))
	if err != nil {
		return fmt.Errorf("Error running connection boilerplate: %s", err)
	}

	log.Debug(dockerHost)

	fmt.Printf("--tlsverify\n--tlscacert=%q\n--tlscert=%q\n--tlskey=%q\n-H=%s\n",
		authOptions.CaCertPath, authOptions.ClientCertPath, authOptions.ClientKeyPath, dockerHost)

	return nil
}
