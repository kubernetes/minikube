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
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/serviceaction"
)

type SystemdProvisioner struct {
	GenericProvisioner
}

func (p *SystemdProvisioner) String() string {
	return "redhat"
}

func NewSystemdProvisioner(osReleaseID string, d drivers.Driver) SystemdProvisioner {
	return SystemdProvisioner{
		GenericProvisioner{
			Commander:         d,
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/systemd/system/docker.service.d/10-machine.conf",
			OsReleaseID:       osReleaseID,
			// x7NOTE: minikube doesn't want packages...
			// maybe to support the assumption that container/iso has already
			// everything in place
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

func (p *SystemdProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	reloadDaemon := false
	switch action {
	case serviceaction.Start, serviceaction.Restart:
		reloadDaemon = true
	}

	// systemd needs reloaded when config changes on disk; we cannot
	// be sure exactly when it changes from the provisioner so
	// we call a reload on every restart to be safe
	if reloadDaemon {
		if _, err := p.RunCmd(exec.Command("sudo", "systemctl", "daemon-reload")); err != nil {
			return err
		}
	}

	if _, err := p.RunCmd(exec.Command("sudo", "systemctl", "-f", action.String(), name)); err != nil {
		return err
	}

	return nil
}
