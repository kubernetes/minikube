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
	"k8s.io/minikube/pkg/minikube/preload"
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
// k8sVersion is used to name the volume and baseImage is the image that is run
// to extract the preloaded images to the volume
func CreatePreloadedImagesVolume(k8sVersion, cRuntime, baseImage, profile string) (string, error) {
	if cRuntime != "docker" {
		return "", nil
	}
	if err := PointToHostDockerDaemon(); err != nil {
		return "", errors.Wrap(err, "point host docker-daemon")
	}
	volumeName := preloadedVolumeName(k8sVersion, profile)
	if dockerVolumeExists(volumeName) {
		return volumeName, nil
	}
	if err := createDockerVolume(volumeName); err != nil {
		return "", errors.Wrap(err, "creating docker volume")
	}
	tarballPath := preload.TarballFilepath(k8sVersion)

	if err := extractTarballToVolume(tarballPath, volumeName, baseImage); err != nil {
		// If the extraction didn't work, delete the corrupt docker volume
		if err := deleteDockerVolume(volumeName); err != nil {
			glog.Warningf("Corrupt docker volume %s was not deleted successfully. You may need to delete it manually via `docker volume rm %s` for minikube to continue to work.", volumeName, volumeName)
		}
		return "", errors.Wrap(err, "extracting tarball to volume")
	}
	return volumeName, nil
}

// dockerVolumeExists returns true if a docker volume with the passed in name exists
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
		if strings.TrimSpace(n) == name {
			return true
		}
	}
	return false
}

// extractTarballToVolume runs a docker image imageName which extracts the tarball at tarballPath
// to the volume named volumeName
func extractTarballToVolume(tarballPath, volumeName, imageName string) error {
	if err := PointToHostDockerDaemon(); err != nil {
		return errors.Wrap(err, "point host docker-daemon")
	}
	cmd := exec.Command(Docker, "run", "--rm", "--entrypoint", "/usr/bin/tar", "-v", fmt.Sprintf("%s:/preloaded.tar:ro", tarballPath), "-v", fmt.Sprintf("%s:/extractDir", volumeName), imageName, "-I", "lz4", "-xvf", "/preloaded.tar", "-C", "/extractDir")
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

// deleteDockerVolume deletes a docker volume with the given name
func deleteDockerVolume(name string) error {
	if err := PointToHostDockerDaemon(); err != nil {
		return errors.Wrap(err, "point host docker-daemon")
	}
	cmd := exec.Command(Docker, "volume", "rm", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "output %s", string(out))
	}
	return nil
}

func preloadedVolumeName(k8sVersion, profile string) string {
	return fmt.Sprintf("k8s-%s-%s", k8sVersion, profile)
}

// PreloadedVolumeAttached returns true if the preloaded volume is attached
// to the running profile
func PreloadedVolumeAttached(k8sVersion, profile string) bool {
	glog.Infof("Checking if preloaded volume is attached to %s", profile)
	if err := PointToHostDockerDaemon(); err != nil {
		glog.Infof("error pointing host to docker daemon: %v", err)
		return false
	}
	volumeName := preloadedVolumeName(k8sVersion, profile)
	cmd := exec.Command(Docker, "inspect", "-f", "{{range .Mounts}} {{.Name}} {{end}}", profile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.Infof("error inspecting mounted volumes: %v", err)
		return false
	}
	vols := strings.Split(string(out), " ")
	for _, v := range vols {
		if strings.TrimSpace(v) == volumeName {
			return true
		}
	}
	return false
}
