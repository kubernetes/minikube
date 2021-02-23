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

package kubeadm

import (
	"errors"
	"fmt"
)

// max minutes wait for kubeadm init. usually finishes in less than 1 minute.
// giving it a generous timeout for possible super slow machines.
const initTimeoutMinutes = 10

// max seconds to wait for running kubectl apply manifests to the cluster to exit
const applyTimeoutSeconds = 10

// FailFastError type is an error that could not be solved by trying again
type FailFastError struct {
	Err error
}

func (f *FailFastError) Error() string {
	return f.Err.Error()
}

// ErrNoExecLinux is thrown on linux when the kubeadm binaries are mounted in a noexec volume on Linux as seen in
// https://github.com/kubernetes/minikube/issues/8327#issuecomment-651288459
// this error could be seen on docker/podman or native driver.
var ErrNoExecLinux = &FailFastError{errors.New("mounted kubeadm binary is not executable")}

// ErrInitTimedout is thrown if kubeadm init takes longer than max time allowed
var ErrInitTimedout = fmt.Errorf("kubeadm init timed out in %d minutes", initTimeoutMinutes)
