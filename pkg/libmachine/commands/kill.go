package commands

import "k8s.io/minikube/pkg/libmachine/libmachine"

func cmdKill(c CommandLine, api libmachine.API) error {
	return runAction("kill", c, api)
}
