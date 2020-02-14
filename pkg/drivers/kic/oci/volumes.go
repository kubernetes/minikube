/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"os/exec"

	"github.com/pkg/errors"
)

// DeleteAllVolumesByLabel delets all volumes that have a specific label
// example: docker volume prune -f --filter label=name.minikube.sigs.k8s.io=minikube
func DeleteAllVolumesByLabel(ociBin string, label string) error {
	if ociBin == Docker {
		if err := PointToHostDockerDaemon(); err != nil {
			return errors.Wrap(err, "point host docker-daemon")
		}
	}
	cmd := exec.Command(ociBin, "volume", "prune", "-f", "--filter", label)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "output %s", string(out))
	}
	return nil
}

// createDockerVolume creates a docker volume to be attached to the container with correct labels and prefixes based on profile name
// Caution ! if volume already exists does NOT return an error and will not apply the minikube labels on it.
// TODO: this should be fixed as a part of https://github.com/kubernetes/minikube/issues/6530
func createDockerVolume(name string) error {
	if err := PointToHostDockerDaemon(); err != nil {
		return errors.Wrap(err, "point host docker-daemon")
	}
	cmd := exec.Command(Docker, "volume", "create", name, "--label", "name.minikube.sigs.k8s.io="+name, "--label", "craeted_by_minikube.minikube.sigs.k8s.io=true")
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "output %s", string(out))
	}
	return nil
}
