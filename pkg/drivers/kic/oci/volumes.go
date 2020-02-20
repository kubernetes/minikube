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
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// DeleteAllVolumesByLabel deletes all volumes that have a specific label
// if there is no volume to delete it will return nil
// example: docker volume prune -f --filter label=name.minikube.sigs.k8s.io=minikube
func DeleteAllVolumesByLabel(ociBin string, label string) error {
	glog.Infof("trying to prune all %s volumes with label %s", ociBin, label)
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

func CreatePreloadedImagesVolume(kicVersion, k8sVersion string) (string, error) {
	if err := PointToHostDockerDaemon(); err != nil {
		return "", errors.Wrap(err, "point host docker-daemon")
	}
	volumeName := fmt.Sprintf("%s-k8s-%s", kicVersion, k8sVersion)
	if dockerVolumeExists(volumeName) {
		return volumeName, nil
	}
	if err := createDockerVolume(volumeName); err != nil {
		return "", errors.Wrap(err, "creating docker volume")
	}
	targetDir := localpath.MakeMiniPath("cache", "preloaded-tarball")
	tarballPath := path.Join(targetDir, fmt.Sprintf("%s-k8s-%s.tar", kicVersion, k8sVersion))

	if err := extractTarballToVolume(tarballPath, volumeName); err != nil {
		return "", errors.Wrap(err, "extracting tarball to volume")
	}
	return volumeName, nil
}

func dockerVolumeExists(name string) bool {
	if err := PointToHostDockerDaemon(); err != nil {
		return false
	}
	cmd := exec.Command(Docker, "volume", "ls", "-q")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	names := strings.Split(string(out), "\n")
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

func extractTarballToVolume(tarballPath, volumeName string) error {
	if err := PointToHostDockerDaemon(); err != nil {
		return errors.Wrap(err, "point host docker-daemon")
	}
	cmd := exec.Command(Docker, "-v", fmt.Sprintf("%s:/preloaded.tar:ro", tarballPath), "-v", fmt.Sprintf("%s:/extractDir", volumeName), "busybox", "tar", "xvf", "/preloaded.tar", "-C", "/extractDir")
	fmt.Println(cmd.Args)
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
