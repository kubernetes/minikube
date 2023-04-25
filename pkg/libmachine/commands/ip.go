package commands

import "k8s.io/minikube/pkg/libmachine/libmachine"

func cmdIP(c CommandLine, api libmachine.API) error {
	return runAction("ip", c, api)
}
