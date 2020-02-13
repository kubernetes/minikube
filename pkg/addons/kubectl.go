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

package addons

import (
	"fmt"
	"os/exec"
	"path"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

var (
	// For testing
	k8sVersion = kubernetesVersion
)

func kubectlCommand(profile string, files []string, enable bool) (*exec.Cmd, error) {
	v, err := k8sVersion(profile)
	if err != nil {
		return nil, err
	}
	kubectlBinary := kubectlBinaryPath(v)

	kubectlAction := "apply"
	if !enable {
		kubectlAction = "delete"
	}

	args := []string{fmt.Sprintf("KUBECONFIG=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")), kubectlBinary, kubectlAction}
	for _, f := range files {
		args = append(args, []string{"-f", f}...)
	}

	cmd := exec.Command("sudo", args...)
	return cmd, nil
}

func kubernetesVersion(profile string) (string, error) {
	cc, err := config.Load(profile)
	if err != nil && !config.IsNotExist(err) {
		return "", err
	}
	version := constants.DefaultKubernetesVersion
	if cc != nil {
		version = cc.KubernetesConfig.KubernetesVersion
	}
	return version, nil
}

func kubectlBinaryPath(version string) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", version, "kubectl")
}
