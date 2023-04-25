package commands

import "k8s.io/minikube/pkg/libmachine/libmachine"

func cmdUpgrade(c CommandLine, api libmachine.API) error {
	return runAction("upgrade", c, api)
}
