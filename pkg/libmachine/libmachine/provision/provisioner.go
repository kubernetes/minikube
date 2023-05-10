/*
Copyright 2023 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provision

import (
	"fmt"
	"os/exec"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/serviceaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
)

var (
	provisioners                     = make(map[string]*RegisteredProvisioner)
	specialCaseProvisioners          = make(map[string]*RegisteredProvisioner)
	detector                Detector = &StandardDetector{}
)

const (
	LastReleaseBeforeCEVersioning = "1.13.1"
)

type Commander interface {
	// RunCmd is just a short-hand for accessing the Runner from the driver.
	RunCmd(cmd *exec.Cmd) (*runner.RunResult, error)
}

type Detector interface {
	DetectProvisioner(d drivers.Driver) (Provisioner, error)
}

// StandardDetector is used by default (for non-testing purposes)
type StandardDetector struct{}

// SetDetector is used to set the detector for anything non-standard..
// like for testing purposes
func SetDetector(newDetector Detector) {
	detector = newDetector
}

// Provisioner defines distribution specific actions
type Provisioner interface {
	fmt.Stringer
	Commander

	// x7NOTE: GenerateDockerOptions has moved inside the cruntimeInstaller
	// check this:
	// git show master:pkg/provision/ubuntu.go
	// (we're using it for container-based machines)
	// In this provisioner and the buildroot one (for iso-based machines)
	// the GenerateDockerOptions returns (DockerOptions, error);
	// we're _'ing the DockerOptions,
	// the error is given by updateUnit(...)
	// Basicallt updating the .service file inside the machine.
	// i.e. We're already making use of the DockerOptions without needing
	// the DockerOptions struct

	// Get the directory where the settings files for docker are to be found
	GetDockerOptionsDir() string

	// Return the auth options used to configure remote connection for the daemon.
	GetAuthOptions() auth.Options

	// Get the swarm options associated with this host.
	GetSwarmOptions() swarm.Options

	// Run a package action e.g. install
	Package(name string, action pkgaction.PackageAction) error

	// Get Hostname
	Hostname() (string, error)

	// Set hostname
	SetHostname(hostname string) error

	// Figure out if this is the right provisioner to use based on /etc/os-release info
	CompatibleWithHost() bool

	// Do the actual provisioning piece:
	//     1. Set the hostname on the instance.
	//     2. Install Docker if it is not present.
	//     3. Configure the daemon to accept connections over TLS.
	//     4. Copy the needed certificates to the server and local config dir.
	//     5. Configure / activate swarm if applicable.
	Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error

	// Perform action on a named service e.g. stop
	Service(name string, action serviceaction.ServiceAction) error

	// Get the driver which is contained in the provisioner.
	GetDriver() drivers.Driver

	// Set the OS Release info depending on how it's represented
	// internally
	SetOsReleaseInfo(info *OsRelease)

	// Get the OS Release info for the current provisioner
	GetOsReleaseInfo() (*OsRelease, error)
}

// RegisteredProvisioner's purpose is to give a representation of an initialized provisioner.
// we pair this struct to a provisioner name into a global var that keeps track
// of the usable provisioners
type RegisteredProvisioner struct {
	New func(d drivers.Driver) Provisioner
}

// Register is used to mark provisioners as "usable" within the codebase,
// as when detecting provisioners we're iterating over this global var
func Register(name string, p *RegisteredProvisioner) {
	provisioners[name] = p
}

// RegisterSpecial is just like Register, only it does register container/iso provisioners
func RegisterSpecial(name string, p *RegisteredProvisioner) {
	specialCaseProvisioners[name] = p
}

// DetectProvisioner is the entrypoint for the provision package.
func DetectProvisioner(d drivers.Driver) (Provisioner, error) {
	return detector.DetectProvisioner(d)
}

// DetectProvisioner is used if detector global var is not overridden for testing ecc...
// Its purpose is to return a compatible provisioner for the machine, given a driver.
// It takes care of the "special-case" machines (container/iso)
func (detector StandardDetector) DetectProvisioner(d drivers.Driver) (Provisioner, error) {
	klog.Infof("Detecting the provisioner...")

	// we're using the provisioner in cache, if any
	if provisioner := GetCachedProvisioner(); provisioner != nil {
		klog.Infof("using cached provisioner")
		return provisioner, nil
	}

	// we're checking special-case machines
	if d.IsISOBased() {
		klog.Infof("detected iso-based machine driver")
		p := specialCaseProvisioners["buildroot"].New(d)
		SetCachedProvisioner(p)
		return p, nil
	} else if d.IsContainerBased() {
		klog.Infof("detected container-based machine driver")
		p := specialCaseProvisioners["container"].New(d)
		SetCachedProvisioner(p)
		return p, nil
	}
	// everything else is a local/remote managed/unmanaged vm

	// we're checking the os-release to test the provisioner for compatibility
	osReleaseOut, err := d.RunCmd(exec.Command("cat", "/etc/os-release"))
	if err != nil {
		return nil, fmt.Errorf("Error running command inside machine, while detecting provisioner: %s", err)
	}

	osReleaseInfo, err := NewOsRelease(osReleaseOut.Stdout.Bytes())
	if err != nil {
		return nil, fmt.Errorf("Error parsing /etc/os-release file: %s", err)
	}

	// we initialize all registered provisioners until we found one compatible with
	// the detected os-release inside the machine.
	// Once found, we cache it for later use..
	for _, p := range provisioners {
		provisioner := p.New(d)
		klog.Infof("we're trying provisioner: %s", provisioner.String())
		provisioner.SetOsReleaseInfo(osReleaseInfo)

		if provisioner.CompatibleWithHost() {
			klog.Infof("found compatible host for ID: %s", osReleaseInfo.ID)
			SetCachedProvisioner(provisioner)
			return provisioner, nil
		}
	}

	// if nothing is found, we return an error
	return nil, ErrDetectionFailed
}
