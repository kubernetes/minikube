package provider

import "k8s.io/minikube/pkg/libmachine/libmachine/host"

type Provider interface {
	// IsValid checks whether or not the Provider can successfully create
	// machines.  If the check does not pass, the provider is no good.
	IsValid() bool

	// Create calls out to the driver this provider is associated with, to
	// actually create the resource.
	Create() (host.Host, error)
}
