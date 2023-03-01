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
	"context"
	"fmt"
	"os/exec"
	"path"

	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

func kubectlCommand(ctx context.Context, cc *config.ClusterConfig, files []string, enable, force bool) *exec.Cmd {
	v := constants.DefaultKubernetesVersion
	if cc != nil {
		v = cc.KubernetesConfig.KubernetesVersion
	}

	kubectlBinary := kapi.KubectlBinaryPath(v)

	kubectlAction := "apply"
	if !enable {
		kubectlAction = "delete"
	}

	args := []string{fmt.Sprintf("KUBECONFIG=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")), kubectlBinary, kubectlAction}
	if force {
		args = append(args, "--force")
	}
	if !enable {
		// --ignore-not-found just ignores when we try to delete a resource that is already gone,
		// like a completed job with a ttlSecondsAfterFinished
		args = append(args, "--ignore-not-found")
	}
	for _, f := range files {
		args = append(args, []string{"-f", f}...)
	}

	return exec.CommandContext(ctx, "sudo", args...)
}
