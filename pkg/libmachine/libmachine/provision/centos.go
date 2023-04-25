package provision

import (
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
)

func init() {
	Register("Centos", &RegisteredProvisioner{
		New: NewCentosProvisioner,
	})
}

func NewCentosProvisioner(d drivers.Driver) Provisioner {
	return &CentosProvisioner{
		NewRedHatProvisioner("centos", d),
	}
}

type CentosProvisioner struct {
	*RedHatProvisioner
}

func (provisioner *CentosProvisioner) String() string {
	return "centos"
}
