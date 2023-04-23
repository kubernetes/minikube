package machine

import "k8s.io/minikube/pkg/libmachine/drivers"

type V2 struct {
	ConfigVersion  int
	Driver         drivers.Driver
	DriverName     string
	MachineOptions *Options
	Name           string
}
