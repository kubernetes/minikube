// +build !darwin

package vmwarefusion

import "k8s.io/minikube/pkg/libmachine/libmachine/drivers"

func NewDriver(hostName, storePath string) drivers.Driver {
	return drivers.NewDriverNotSupported("vmwarefusion", hostName, storePath)
}
