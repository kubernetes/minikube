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
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"k8s.io/klog/v2"
)

// RemoveVolume removes a volume
func RemoveVolume(ociBin string, name string) error {
	if !volumeExists(ociBin, name) {
		return nil
	}
	rr, err := runCmd(exec.Command(ociBin, "volume", "rm", name))
	if err != nil {
		if strings.Contains(rr.Output(), "No such volume") ||
			strings.Contains(rr.Output(), "no such volume") {
			return ErrVolumeNotFound
		}
	}

	return err
}

func volumeExists(ociBin string, name string) bool {
	_, err := containerVolumeInspect(ociBin, name)
	if err != nil && !errors.Is(err, ErrVolumeNotFound) { // log unexpected error
		klog.Warningf("Error inspecting docker volume %s: %v", name, err)
	}
	return err == nil
}

func containerVolumeInspect(ociBin string, name string) (interface{}, error) {
	var info interface{}
	cmd := exec.Command(ociBin, "volume", "inspect", name)
	rr, err := runCmd(cmd)
	if err != nil {
		if strings.Contains(rr.Output(), "No such volume") ||
			strings.Contains(rr.Output(), "no such volume") {
			return info, ErrVolumeNotFound
		}
		return info, err
	}

	err = json.Unmarshal(rr.Stdout.Bytes(), &info)

	return info, err
}

// DeleteAllVolumesByLabel deletes all volumes that have a specific label
// if there is no volume to delete it will return nil
func DeleteAllVolumesByLabel(ctx context.Context, ociBin string, label string, warnSlow ...bool) []error {
	var deleteErrs []error
	klog.Infof("trying to delete all %s volumes with label %s", ociBin, label)

	vs, err := allVolumesByLabel(ociBin, label)

	if err != nil {
		return []error{fmt.Errorf("listing volumes by label %q: %v", label, err)}
	}

	for _, v := range vs {
		if _, err := runCmd(exec.CommandContext(ctx, ociBin, "volume", "rm", "--force", v), warnSlow...); err != nil {
			deleteErrs = append(deleteErrs, fmt.Errorf("deleting %q", v))
		}
	}

	return deleteErrs
}

// PruneAllVolumesByLabel deletes all volumes that have a specific label
// if there is no volume to delete it will return nil
// example: docker volume prune -f --filter label=name.minikube.sigs.k8s.io=minikube
func PruneAllVolumesByLabel(ctx context.Context, ociBin string, label string, warnSlow ...bool) []error {
	var deleteErrs []error
	klog.Infof("trying to prune all %s volumes with label %s", ociBin, label)
	cmd := exec.CommandContext(ctx, ociBin, "volume", "prune", "-f", "--filter", "label="+label)
	if _, err := runCmd(cmd, warnSlow...); err != nil {
		deleteErrs = append(deleteErrs, errors.Wrapf(err, "prune volume by label %s", label))
	}

	return deleteErrs
}

// allVolumesByLabel returns name of all docker volumes by a specific label
// will not return error if there is no volume found.
func allVolumesByLabel(ociBin string, label string) ([]string, error) {
	rr, err := runCmd(exec.Command(ociBin, "volume", "ls", "--filter", "label="+label, "--format", "{{.Name}}"))
	s := bufio.NewScanner(bytes.NewReader(rr.Stdout.Bytes()))
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
func ExtractTarballToVolume(ociBin string, tarballPath, volumeName, imageName string) error {
	cmdArgs := []string{"run", "--rm", "--entrypoint", "/usr/bin/tar"}
	// Podman:
	// when selinux setenforce is enforced, normal mount will lead to file permissions error (-?????????)
	// - option 1: label the file as container private (mount option :Z), but will alter the file in the host machine
	// - option 2*: keep the file untouched and set --security-opt label=disable (no changes to file)
	if ociBin == Podman && runtime.GOOS == "linux" {
		cmdArgs = append(cmdArgs, "--security-opt", "label=disable")
	}
	cmdArgs = append(cmdArgs, "-v", fmt.Sprintf("%s:/preloaded.tar:ro", tarballPath), "-v", fmt.Sprintf("%s:/extractDir", volumeName), imageName, "-I", "lz4", "-xf", "/preloaded.tar", "-C", "/extractDir")
	cmd := exec.Command(ociBin, cmdArgs...)
	if _, err := runCmd(cmd); err != nil {
		return err
	}
	return nil
}

// createVolume creates a volume to be attached to the container with correct labels and prefixes based on profile name
// Caution ! if volume already exists does NOT return an error and will not apply the minikube labels on it.
// TODO: this should be fixed as a part of https://github.com/kubernetes/minikube/issues/6530
func createVolume(ociBin string, profile string, nodeName string) error {
	if _, err := runCmd(exec.Command(ociBin, "volume", "create", nodeName, "--label", fmt.Sprintf("%s=%s", ProfileLabelKey, profile), "--label", fmt.Sprintf("%s=%s", CreatedByLabelKey, "true"))); err != nil {
		return err
	}
	return nil
}

// prepareVolumeSideCar will copy the initial content of the mount point by starting a temporary container to check the expected content
func prepareVolumeSideCar(ociBin string, imageName string, nodeName string) error {
	cmdArgs := []string{"run", "--rm", "--name", fmt.Sprintf("%s-preload-sidecar", nodeName), "--entrypoint", "/usr/bin/test"}
	cmdArgs = append(cmdArgs, "-v", fmt.Sprintf("%s:/var", nodeName), imageName, "-d", "/var/lib")
	cmd := exec.Command(ociBin, cmdArgs...)
	if _, err := runCmd(cmd); err != nil {
		return err
	}
	return nil
}
