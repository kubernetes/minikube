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

	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
)

type GenericProvisioner struct {
	OsReleaseID       string
	DockerOptionsDir  string
	DaemonOptionsFile string
	Packages          []string
	OsReleaseInfo     *OsRelease
	Driver            drivers.Driver
	AuthOptions       *auth.Options
	EngineOptions     *engine.Options
	SwarmOptions      swarm.Options
}

func (provisioner *GenericProvisioner) RunCmd(cmd *exec.Cmd) (*runner.RunResult, error) {
	return provisioner.GetDriver().RunCmd(cmd)
}

func (provisioner *GenericProvisioner) Hostname() (string, error) {
	rr, err := provisioner.RunCmd(exec.Command("hostname"))
	if err != nil {
		return "", err
	}

	return rr.Stdout.String(), nil
}

func (provisioner *GenericProvisioner) SetHostname(hostname string) error {
	cmd := fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /etc/hostname",
		hostname,
		hostname,
	)

	if _, err := provisioner.RunCmd(exec.Command("bash", "-c", cmd)); err != nil {
		return err
	}

	// ubuntu/debian use 127.0.1.1 for non "localhost" loopback hostnames: https://www.debian.org/doc/manuals/debian-reference/ch05.en.html#_the_hostname_resolution
	cmd = fmt.Sprintf(`
		if ! grep -xq '.*\s%s' /etc/hosts; then
			if grep -xq '127.0.1.1\s.*' /etc/hosts; then
				sudo sed -i 's/^127.0.1.1\s.*/127.0.1.1 %s/g' /etc/hosts;
			else 
				echo '127.0.1.1 %s' | sudo tee -a /etc/hosts; 
			fi
		fi`,
		hostname,
		hostname,
		hostname,
	)
	if _, err := provisioner.RunCmd(exec.Command("bash", "-c", cmd)); err != nil {
		return err
	}

	return nil
}

func (provisioner *GenericProvisioner) GetDockerOptionsDir() string {
	return provisioner.DockerOptionsDir
}

func (provisioner *GenericProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID
}

func (provisioner *GenericProvisioner) GetAuthOptions() auth.Options {
	return *provisioner.AuthOptions
}

func (provisioner *GenericProvisioner) GetSwarmOptions() swarm.Options {
	return provisioner.SwarmOptions
}

func (provisioner *GenericProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}

func (provisioner *GenericProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return provisioner.OsReleaseInfo, nil
}

func (provisioner *GenericProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
