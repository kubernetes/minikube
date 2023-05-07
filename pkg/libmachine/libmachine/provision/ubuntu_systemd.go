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
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/cruntimeInstaller"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
)

func init() {
	Register("Ubuntu-SystemD", &RegisteredProvisioner{
		New: NewUbuntuSystemdProvisioner,
	})
}

func NewUbuntuSystemdProvisioner(d drivers.Driver) Provisioner {
	return &UbuntuSystemdProvisioner{
		NewSystemdProvisioner("ubuntu", d),
	}
}

type UbuntuSystemdProvisioner struct {
	SystemdProvisioner
}

func (provisioner *UbuntuSystemdProvisioner) String() string {
	return "ubuntu(systemd)"
}

func (provisioner *UbuntuSystemdProvisioner) CompatibleWithHost() bool {
	const FirstUbuntuSystemdVersion = 15.04

	isUbuntu := provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID
	if !isUbuntu {
		return false
	}
	versionNumber, err := strconv.ParseFloat(provisioner.OsReleaseInfo.VersionID, 64)
	if err != nil {
		return false
	}
	return versionNumber >= FirstUbuntuSystemdVersion

}

func (provisioner *UbuntuSystemdProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	updateMetadata := true

	switch action {
	case pkgaction.Install, pkgaction.Upgrade:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
		updateMetadata = false
	case pkgaction.Purge:
		packageAction = "purge"
		updateMetadata = false
	}

	switch name {
	case "docker":
		name = "docker-ce"
	}

	if updateMetadata {
		if err := waitForLockAptGetUpdate(provisioner); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive sudo -E apt-get %s -y  %s", packageAction, name)

	klog.Infof("package: action=%s name=%s", action.String(), name)

	return waitForLock(provisioner, exec.Command("bash", "-c", command))
}

func (provisioner *UbuntuSystemdProvisioner) dockerDaemonResponding() bool {
	klog.Infof("checking docker daemon")

	if out, err := provisioner.RunCmd(exec.Command("sudo", "docker", "version")); err != nil {
		klog.Infof("Error running command to check if the daemon is up: %s", err)
		klog.Infof("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *UbuntuSystemdProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	if !provisioner.Driver.IsManaged() {
		return nil
	}

	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = &authOptions
	provisioner.EngineOptions = &engineOptions
	swarmOptions.Env = engineOptions.Env

	storageDriver, err := decideStorageDriver(provisioner, "overlay2", engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	klog.Infof("setting hostname")
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	// x7NOTE: installation of packages is not mandatory. The provisioner explicitly
	// takes care of that inside the Provision method for all implementations.
	// Thus we can safe have the container/iso provisioners embed the libmachine's
	// systemdProvisioner (that has packages within.. as opposed to minikube's codebase
	// systemdProvisioner wich is the same except for the packages), and forget those
	// inside the actual Provision() call.
	klog.Infof("installing base packages")
	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	klog.Infof("configuring auth")
	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	klog.Infof("installing container runtime into the machine")

	rnr, err := provisioner.Driver.GetRunner()
	if err != nil {
		return errors.Wrap(err, "while getting runner for cruntime installer")
	}

	instllr := cruntimeInstaller.DetectCRuntimeInstaller(provisioner.EngineOptions, rnr, provisioner.Driver.DriverName(), provisioner.AuthOptions)
	return instllr.InstallCRuntime()
}
