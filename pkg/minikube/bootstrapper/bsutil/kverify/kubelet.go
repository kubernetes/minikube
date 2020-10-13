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

// Package kverify verifies a running Kubernetes cluster is healthy
package kverify

import (
	"fmt"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/util/retry"
)

// KubeletStatus checks the kubelet status
func KubeletStatus(cr command.Runner) state.State {
	klog.Infof("Checking kubelet status ...")
	active := sysinit.New(cr).Active("kubelet")
	if active {
		return state.Running
	}
	return state.Stopped
}

// WaitForKubelet will wait for Kubelet service to be running ...
func WaitForKubelet(cr command.Runner, timeout time.Duration) error {
	pStart := time.Now()
	klog.Infof("waiting for kubelet to be running ....")
	kr := func() error {
		if st := KubeletStatus(cr); st != state.Running {
			return fmt.Errorf("status %s", st)
		}
		return nil
	}

	if err := retry.Local(kr, timeout); err != nil {
		return fmt.Errorf("not running: %s", err)
	}

	klog.Infof("duration metric: took %s WaitForKubelet to finish.", time.Since(pStart))

	return nil

}
