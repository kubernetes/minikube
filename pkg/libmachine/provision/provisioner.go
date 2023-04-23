package provision

import (
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cruntime"
	"k8s.io/minikube/pkg/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/provision/serviceaction"
)

// Provisioner is the interface that we use in order to "provision" a machine.
// Once the Driver has provisioned a linux machine to run kubernetes on,
// the provisioner prepares it for kubernetes:
// it configures machine's hostname,
// it installs packages, configures services,
// sets the authentication for the container runtime...
type Provisioner interface {
	// GetProvisionerName gets the name of the provisioner
	GetProvisionerName() string
	// CompatibleWithMachine returns true if this provisioner can be used
	// for the machine we're creating, otherwise returns false
	CompatibleWithMachine() bool

	// RunCommand escapes to the machine's shell and runs a command
	// Provides the primitive. Everything else (pkg/service mgmt..)
	// should be built on top of this, by utils functions.
	RunCommand(string) (string, error)
	// ServiceAction tries to set a system service inside the machine, in a given state.
	ServiceAction(string, serviceaction.ServiceAction) error
	// PackageAction tries to perform an action on a package inside the machine (e.g. install/remove/...)
	PackageAction(string, pkgaction.PackageAction) error

	// GetContainerRuntime gets the container runtime that has been chosen to run
	// the kubernetes cluster
	// NOTE: WHO?
	GetContainerRuntime() string
	// GetContainerRuntimeOptionsDir gets the path for the runtime config directory
	// NOTE: WHERE?
	GetContainerRuntimeOptionsDir() string
	// SetContainerRuntimeOptions sets the configuration for the chosen container runtime
	// NOTE: WHAT?
	GenerateContainerRuntimeOptions(interface{}) (*ContainerRuntimeOptions, error)

	// GetAuthOptions returns the authentication options used to configure the remote
	// connection for the container runtime daemon
	GetAuthOptions() auth.Options

	// GetOsReleaseInfo gets os related info from the linux machine
	GetOsReleaseInfo() (*OsRelease, error)
	// SetOsReleaseInfo sets os related info values in the provisioner struct
	SetOsReleaseInfo(*OsRelease)

	// Provision does the actual provisioning piece:
	//     1. Set the hostname on the machine.
	//     2. Install the chosen container runtime if not already present.
	//     3. Configure the daemon to accept connections over TLS.
	//     4. Copy the needed certificates to the server and local config dir.
	//     5. Configure / activate swarm if applicable.
	Provision(auth.Options, cruntime.Options) error
}
