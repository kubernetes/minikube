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

// ServiceStatus checks the status of a systemd or init.d service
func ServiceStatus(cr command.Runner, svc string) state.State {
	active := sysinit.New(cr).Active(svc)
	if active {
		return state.Running
	}
	return state.Stopped
}

// WaitForService will wait for a "systemd" or "init.d" service to be running on the node...
// not to be confused with Kubernetes Services
func WaitForService(cr command.Runner, svc string, timeout time.Duration) error {
	pStart := time.Now()
	klog.Infof("waiting for %s service to be running ....", svc)
	kr := func() error {
		if st := ServiceStatus(cr, svc); st != state.Running {
			return fmt.Errorf("status %s", st)
		}
		return nil
	}

	if err := retry.Local(kr, timeout); err != nil {
		return fmt.Errorf("not running: %s", err)
	}

	klog.Infof("duration metric: took %s WaitForService to wait for %s.", time.Since(pStart), svc)

	return nil

}
