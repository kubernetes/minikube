/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package pause

import (
	"os/exec"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
)

const pausedFile = "paused"

// CreatePausedFile creates a file in the minikube cluster to indicate that the apiserver is paused
func CreatePausedFile(r command.Runner) {
	if _, err := r.RunCmd(exec.Command("touch", pausedFile)); err != nil {
		klog.Errorf("failed to create paused file, apiserver may display incorrect status")
	}
}

// RemovePausedFile removes a file in minikube cluster to indicate that the apiserver is unpaused
func RemovePausedFile(r command.Runner) {
	if _, err := r.RunCmd(exec.Command("rm", "-f", pausedFile)); err != nil {
		klog.Errorf("failed to remove paused file, apiserver may display incorrect status")
	}
}
