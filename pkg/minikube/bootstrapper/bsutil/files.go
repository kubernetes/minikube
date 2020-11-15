/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

// Package bsutil will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"sort"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// KubeadmYamlPath is the path to the kubeadm configuration
var KubeadmYamlPath = path.Join(vmpath.GuestEphemeralDir, "kubeadm.yaml")

const (
	// KubeletServiceFile is the file for the systemd kubelet.service
	KubeletServiceFile = "/lib/systemd/system/kubelet.service"
	// KubeletSystemdConfFile is config for the systemd kubelet.service
	KubeletSystemdConfFile = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	// InitRestartWrapper is ...
	InitRestartWrapper = "/etc/init.d/.restart_wrapper.sh"
	// KubeletInitPath is where Sys-V style init script is installed
	KubeletInitPath = "/etc/init.d/kubelet"
)

// CopyFiles combines mkdir requests into a single call to reduce load
func CopyFiles(runner command.Runner, files []assets.CopyableFile) error {
	dirs := []string{}
	for _, f := range files {
		dirs = append(dirs, f.GetTargetDir())
	}
	args := append([]string{"mkdir", "-p"}, dirs...)
	if _, err := runner.RunCmd(exec.Command("sudo", args...)); err != nil {
		return errors.Wrap(err, "mkdir")
	}

	for _, f := range files {
		if err := runner.Copy(f); err != nil {
			return errors.Wrapf(err, "copy")
		}
	}
	return nil
}

// ReverseDirList returns a list of subdirectories under the given path, sorted in reverse order, and any error.
// If n > 0, ReverseDirList returns at most n subdirectories.
// If n <= 0, ReverseDirList returns all the subdirectories from the directory.
func ReverseDirList(path string, n int) (list []string, err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to list directory %s: %w", path, err)
	}
	for _, file := range files {
		if file.IsDir() {
			list = append(list, file.Name())
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i] > list[j] })
	if n <= 0 || n > len(list) {
		n = len(list)
	}
	return list[0:n], nil
}
