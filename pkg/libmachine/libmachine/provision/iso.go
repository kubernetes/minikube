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
	RegisterSpecial("buildroot", &RegisteredProvisioner{
		New: NewBuildrootProvisioner,
	})
}

// BuildrootProvisioner provisions the custom system based on Buildroot
type BuildrootProvisioner struct {
	SystemdProvisioner
	clusterName string
}

// NewBuildrootProvisioner creates a new BuildrootProvisioner
func NewBuildrootProvisioner(d drivers.Driver) Provisioner {
	return &BuildrootProvisioner{
		NewSystemdProvisioner("buildroot", d),
		viper.GetString(config.ProfileName),
	}
}

func (p *BuildrootProvisioner) String() string {
	return "buildroot"
}

// CompatibleWithHost checks if provisioner is compatible with host
func (p *BuildrootProvisioner) CompatibleWithHost() bool {
	return p.OsReleaseInfo.ID == "buildroot"
}

// Package installs a package
func (p *BuildrootProvisioner) Package(_ string, _ pkgaction.PackageAction) error {
	return nil
}

// Provision does the provisioning
func (p *BuildrootProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
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

	klog.Infof("installing container runtime into the machine")

	rnr, err := p.Driver.GetRunner()
	if err != nil {
		return errors.Wrap(err, "while getting runner for cruntime installer")
	}

	instllr := cruntimeInstaller.DetectCRuntimeInstaller(p.EngineOptions, rnr, p.Driver.DriverName(), p.AuthOptions)
	return instllr.InstallCRuntime()
}
