package commands

import "k8s.io/minikube/pkg/libmachine/libmachine"

func cmdProvision(c CommandLine, api libmachine.API) error {
	return runAction("provision", c, api)
}
