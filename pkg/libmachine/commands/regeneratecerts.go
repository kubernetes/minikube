package commands

import (
	"k8s.io/minikube/pkg/libmachine/libmachine"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
)

func cmdRegenerateCerts(c CommandLine, api libmachine.API) error {
	if !c.Bool("force") {
		ok, err := confirmInput("Regenerate TLS machine certs?  Warning: this is irreversible.")
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}
	}

	log.Infof("Regenerating TLS certificates")

	if c.Bool("client-certs") {
		return runAction("configureAllAuth", c, api)
	}
	return runAction("configureAuth", c, api)
}
