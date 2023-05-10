package provision

import (
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/drivers/mockdriver"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/serviceaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
)

// MockProvisioner defines distribution specific actions
type MockProvisioner struct {
	Provisioned bool
}

func (provisioner *MockProvisioner) String() string {
	return "mock"
}

// Service performs an action for a service
func (provisioner *MockProvisioner) Service(_ string, _ serviceaction.ServiceAction) error {
	return nil
}

// Package performs an action for a package
func (provisioner *MockProvisioner) Package(_ string, _ pkgaction.PackageAction) error {
	return nil
}

// Hostname returns the hostname
func (provisioner *MockProvisioner) Hostname() (string, error) {
	return "mockhostname", nil
}

// SetHostname sets the hostname
func (provisioner *MockProvisioner) SetHostname(_ string) error {
	return nil
}

// GetDockerOptionsDir gets Docker options dir
func (provisioner *MockProvisioner) GetDockerOptionsDir() string {
	return "/mockdirectory"
}

// GetAuthOptions returns a the auth.Options
func (provisioner *MockProvisioner) GetAuthOptions() auth.Options {
	return auth.Options{}
}

// CompatibleWithHost checks if provisioner is compatible with host
func (provisioner *MockProvisioner) CompatibleWithHost() bool {
	return true
}

// SetOsReleaseInfo sets the os-release info
func (provisioner *MockProvisioner) SetOsReleaseInfo(_ *OsRelease) {
}

// GetOsReleaseInfo gets the os-release info
func (provisioner *MockProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return &OsRelease{}, nil
}

// AttemptIPContact attempts to contact an IP and port
func (provisioner *MockProvisioner) AttemptIPContact(_ int) {
}

// Provision provisions the machine
func (provisioner *MockProvisioner) Provision(_ swarm.Options, _ auth.Options, _ engine.Options) error {
	provisioner.Provisioned = true
	return nil
}

// SSHCommand runs a SSH command
func (provisioner *MockProvisioner) SSHCommand(_ string) (string, error) {
	return "", nil
}

// GetDriver gets the driver
func (provisioner *MockProvisioner) GetDriver() drivers.Driver {
	return &mockdriver.MockDriver{}
}

// RunCmd mocks a command inside the linux machine
func (provisioner *MockProvisioner) RunCmd(_ *exec.Cmd) (*runner.RunResult, error) {
	return nil, nil
}

// GetSwarmOptions gets the swarm.Options
func (provisioner *MockProvisioner) GetSwarmOptions() swarm.Options {
	return swarm.Options{}
}

// MockDetector can detect MockProvisioner
type MockDetector struct {
	Provisioner *MockProvisioner
}

// DetectProvisioner detects a provisioner
func (m *MockDetector) DetectProvisioner(_ drivers.Driver) (Provisioner, error) {
	return m.Provisioner, nil
}
