package provision

import (
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cruntime"
	"k8s.io/minikube/pkg/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/provision/serviceaction"
	"k8s.io/minikube/pkg/libmachine/runner"
)

// Provisioner is the interface that we use in order to transform a
// freshly created linux machine into a machine ready to host kubernetes.
// In order to do that, the provisioner needs to be able to install packages,
// to start/stop services, to manage configuration files, ...
// we could've used `type Ansible interface` instead...
// The Provisioner is an interface, because the implementation has to vary
// depending on the kind of machine we're provisioning:
// many things can vary: init system, package manager, file location, ...
type Provisioner interface {
	// GetProvisionerName gets the name of the provisioner
	GetProvisionerName() string
	// CompatibleWithMachine returns true if this provisioner can be used
	// for the machine we're creating, otherwise returns false
	CompatibleWithMachine() bool

	// RunCommand escapes to the machine's shell and runs a command
	// Provides the primitive. Everything else (pkg/service mgmt..)
	// should be built on top of this, by utils functions.
	// ...
	// we'll see.. having an embedded runner.Runner interface seems right here...
	// RunCommand(string) (string, error)
	// ..
	runner.Runner
	// ServiceAction tries to set a system service inside the machine, in a given state.
	ServiceAction(string, serviceaction.ServiceAction) error
	// PackageAction tries to perform an action on a package inside the machine (e.g. install/remove/...)
	PackageAction(string, pkgaction.PackageAction) error

	// GetCRuntime gets back a reference to the container runtime
	// that will run kubernetes inside the linux machine
	GetCRuntime() cruntime.CRuntimeEngine

	// GetOsReleaseInfo gets os related info from the linux machine
	GetOsReleaseInfo() (*OsRelease, error)
	// SetOsReleaseInfo sets os related info values in the provisioner struct
	SetOsReleaseInfo(*OsRelease)

	// ConfigureAuth configures the ssl/tls authentication for the container runtime daemon
	// inside the machine
	ConfigureAuth() error
	// Provision does the actual provisioning piece:
	//     1. Set the hostname on the machine.
	//     2. Install the chosen container runtime if not already present.
	//     3. Configure the daemon to accept connections over TLS.
	//     4. Copy the needed certificates to the server and local config dir.
	//     5. Configure / activate swarm if applicable.
	Provision(auth.Options, cruntime.Options) error
}
