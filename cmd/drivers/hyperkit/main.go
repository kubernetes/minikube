package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"k8s.io/minikube/pkg/minikube/drivers/hyperkit"
)

func main() {
	plugin.RegisterDriver(hyperkit.NewDriver("", ""))
}
