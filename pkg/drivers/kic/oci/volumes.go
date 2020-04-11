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
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// DeleteAllVolumesByLabel deletes all volumes that have a specific label
// if there is no volume to delete it will return nil
func DeleteAllVolumesByLabel(ociBin string, label string) []error {
	var deleteErrs []error
	glog.Infof("trying to delete all %s volumes with label %s", ociBin, label)

	vs, err := allVolumesByLabel(ociBin, label)

	if err != nil {
		return []error{fmt.Errorf("listing volumes by label %q: %v", label, err)}
	}

	for _, v := range vs {
		if _, err := WarnIfSlow(ociBin, "volume", "rm", "--force", v); err != nil {
			deleteErrs = append(deleteErrs, fmt.Errorf("deleting %q", v))
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

	if _, err := WarnIfSlow(ociBin, "volume", "prune", "-f", "--filter", "label="+label); err != nil {
		deleteErrs = append(deleteErrs, errors.Wrapf(err, "prune volume by label %s", label))
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

// ExtractTarballToVolume runs a docker image imageName which extracts the tarball at tarballPath
// to the volume named volumeName
func ExtractTarballToVolume(tarballPath, volumeName, imageName string) error {
	cmd := exec.Command(Docker, "run", "--rm", "--entrypoint", "/usr/bin/tar", "-v", fmt.Sprintf("%s:/preloaded.tar:ro", tarballPath), "-v", fmt.Sprintf("%s:/extractDir", volumeName), imageName, "-I", "lz4", "-xvf", "/preloaded.tar", "-C", "/extractDir")
	glog.Infof("executing: %s", cmd.Args)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "output %s", string(out))
	}
	return nil
}

// createDockerVolume creates a docker volume to be attached to the container with correct labels and prefixes based on profile name
// Caution ! if volume already exists does NOT return an error and will not apply the minikube labels on it.
// TODO: this should be fixed as a part of https://github.com/kubernetes/minikube/issues/6530
func createDockerVolume(profile string, nodeName string) error {
	cmd := exec.Command(Docker, "volume", "create", nodeName, "--label", fmt.Sprintf("%s=%s", ProfileLabelKey, profile), "--label", fmt.Sprintf("%s=%s", CreatedByLabelKey, "true"))
	glog.Infof("executing: %s", cmd.Args)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "output %s", string(out))
	}
	return nil
}
