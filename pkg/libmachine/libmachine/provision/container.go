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
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/cruntimeInstaller"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/util/retry"
)

func init() {
	RegisterSpecial("container", &RegisteredProvisioner{
		New: NewContainerProvisioner,
	})
}

// ContainerProvisioner provisions the ubuntu
type ContainerProvisioner struct {
	BuildrootProvisioner
}

// NewContainerProvisioner creates a new ContainerProvisioner
func NewContainerProvisioner(d drivers.Driver) Provisioner {
	return &ContainerProvisioner{
		BuildrootProvisioner{
			NewSystemdProvisioner("container", d),
			viper.GetString(config.ProfileName),
		},
	}
}

func (p *ContainerProvisioner) String() string {
	return "container"
}

// CompatibleWithHost: for lack of a better idea.. this returns true and is the reason for
// the RegisterSpecial function of libmachine/pkg/provision..
// The idea is that we want to say that this provisioner is compatible only if
// we know that we're running inside a container (either podman/docker/...)
func (p *ContainerProvisioner) CompatibleWithHost() bool {
	// here we're assuming that this provisioner has been called correctly
	// based on the driver's property of acting on a container.
	// we also assume that the container is compatible
	// and has all the requirements installed
	return true
}

// Package installs a package
func (p *ContainerProvisioner) Package(_ string, _ pkgaction.PackageAction) error {
	// we assume that all the requirements are already baked inside the container
	return nil
}

// Provision does the provisioning
func (p *ContainerProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	if !p.Driver.IsManaged() {
		return nil
	}

	p.SwarmOptions = swarmOptions
	p.AuthOptions = &authOptions
	p.EngineOptions = &engineOptions

	klog.Infof("provisioning hostname %q", p.Driver.GetMachineName())
	if err := p.SetHostname(p.Driver.GetMachineName()); err != nil {
		return err
	}

	p.AuthOptions = setRemoteAuthOptions(p)
	klog.Infof("set auth options %+v", p.AuthOptions)

	klog.Infof("setting up certificates")
	configAuth := func() error {
		if err := ConfigureAuth(p); err != nil {
			klog.Warningf("configureAuth failed: %v", err)
			return &retry.RetriableError{Err: err}
		}
		return nil
	}

	err := retry.Expo(configAuth, 100*time.Microsecond, 2*time.Minute)

	if err != nil {
		klog.Infof("Error configuring auth during provisioning %v", err)
		return err
	}

	rnr, err := p.Driver.GetRunner()
	if err != nil {
		return errors.Wrap(err, "while getting runner for cruntime installer")
	}

	instllr := cruntimeInstaller.DetectCRuntimeInstaller(p.EngineOptions, rnr, p.Driver.DriverName(), p.AuthOptions)
	instllr.InstallCRuntime()

	return nil
}
