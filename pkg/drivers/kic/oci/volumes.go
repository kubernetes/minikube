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
	"bufio"
	"bytes"
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
func DeleteAllVolumesByLabel(ociBin string, label string) []error {
	var deleteErrs []error
	glog.Infof("trying to delete all %s volumes with label %s", ociBin, label)
	if ociBin == Docker {
		if err := PointToHostDockerDaemon(); err != nil {
			return []error{errors.Wrap(err, "point host docker-daemon")}
		}
	}

	vs, err := allVolumesByLabel(ociBin, label)
	if err != nil {
		return []error{fmt.Errorf("listing volumes by label %q: %v", label, err)}
	}

	for _, v := range vs {
		cmd := exec.Command(ociBin, "volume", "rm", "--force", v)
		if out, err := cmd.CombinedOutput(); err != nil {
			deleteErrs = append(deleteErrs, fmt.Errorf("deleting volume %s: output: %s", v, string(out)))
		}
	}
	return deleteErrs
}

// PruneAllVolumesByLabel deletes all volumes that have a specific label
// if there is no volume to delete it will return nil
// example: docker volume prune -f --filter label=name.minikube.sigs.k8s.io=minikube
func PruneAllVolumesByLabel(ociBin string, label string) []error {
	var deleteErrs []error
	glog.Infof("trying to prune all %s volumes with label %s", ociBin, label)
	if ociBin == Docker {
		if err := PointToHostDockerDaemon(); err != nil {
			return []error{errors.Wrap(err, "point host docker-daemon")}
		}
	}

	// try to prune afterwards just in case delete didn't go through
	cmd := exec.Command(ociBin, "volume", "prune", "-f", "--filter", "label="+label)
	if out, err := cmd.CombinedOutput(); err != nil {
		deleteErrs = append(deleteErrs, errors.Wrapf(err, "prune volume by label %s: %s", label, string(out)))
	}
	return deleteErrs
}

// allVolumesByLabel returns name of all docker volumes by a specific label
// will not return error if there is no volume found.
func allVolumesByLabel(ociBin string, label string) ([]string, error) {
	cmd := exec.Command(ociBin, "volume", "ls", "--filter", "label="+label, "--format", "{{.Name}}")
	stdout, err := cmd.Output()
	s := bufio.NewScanner(bytes.NewReader(stdout))
	var vols []string
	for s.Scan() {
		v := strings.TrimSpace(s.Text())
		if v != "" {
			vols = append(vols, v)
		}
	}
	return vols, err
}

// CreatePreloadedImagesVolume creates a volume with preloaded images
func CreatePreloadedImagesVolume(k8sVersion string) (string, error) {
	if err := PointToHostDockerDaemon(); err != nil {
		return "", errors.Wrap(err, "point host docker-daemon")
	}
	volumeName := fmt.Sprintf("k8s-%s", k8sVersion)
	if dockerVolumeExists(volumeName) {
		return volumeName, nil
	}
	if err := createDockerVolume(volumeName); err != nil {
		return "", errors.Wrap(err, "creating docker volume")
	}
	targetDir := localpath.MakeMiniPath("cache", "preloaded-tarball")
	tarballPath := path.Join(targetDir, fmt.Sprintf("%s.tar", k8sVersion))

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
	cmd := exec.Command(Docker, "run", "-v", fmt.Sprintf("%s:/preloaded.tar:ro", tarballPath), "-v", fmt.Sprintf("%s:/extractDir", volumeName), "busybox", "tar", "xvf", "/preloaded.tar", "-C", "/extractDir")
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
	cmd := exec.Command(Docker, "volume", "create", name, "--label", fmt.Sprintf("%s=%s", ProfileLabelKey, name), "--label", fmt.Sprintf("%s=%s", CreatedByLabelKey, "true"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "output %s", string(out))
	}
	return nil
}
