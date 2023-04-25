package commands

import "k8s.io/minikube/pkg/libmachine/libmachine"

func cmdStop(c CommandLine, api libmachine.API) error {
	return runAction("stop", c, api)
}
