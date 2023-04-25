package host

import "k8s.io/minikube/pkg/libmachine/libmachine/drivers"

type V2 struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *Options
	Name          string
}
