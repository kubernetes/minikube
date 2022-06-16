/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/util/retry"
)

// Pause pauses a Kubernetes cluster, retrying if necessary
func Pause(cr cruntime.Manager, r command.Runner, namespaces []string) ([]string, error) {
	var ids []string
	tryPause := func() (err error) {
		ids, err = pause(cr, r, namespaces)
		return err
	}

	if err := retry.Expo(tryPause, 250*time.Millisecond, 2*time.Second); err != nil {
		return ids, err
	}
	return ids, nil
}

// pause pauses a Kubernetes cluster
func pause(cr cruntime.Manager, r command.Runner, namespaces []string) ([]string, error) {
	ids := []string{}

	// Disable the kubelet so it does not attempt to restart paused pods
	sm := sysinit.New(r)
	klog.Info("kubelet running: ", sm.Active("kubelet"))

	if err := sm.DisableNow("kubelet"); err != nil {
		return ids, errors.Wrap(err, "kubelet disable --now")
	}

	ids, err := cr.ListContainers(cruntime.ListContainersOptions{State: cruntime.Running, Namespaces: namespaces})
	if err != nil {
		return ids, errors.Wrap(err, "list running")
	}

	if len(ids) == 0 {
		klog.Warningf("no running containers to pause")
		return ids, nil
	}

	return ids, cr.PauseContainers(ids)
}

// Unpause unpauses a Kubernetes cluster, retrying if necessary
func Unpause(cr cruntime.Manager, r command.Runner, namespaces []string) ([]string, error) {
	var ids []string
	tryUnpause := func() (err error) {
		ids, err = unpause(cr, r, namespaces)
		return err
	}

	if err := retry.Expo(tryUnpause, 250*time.Millisecond, 2*time.Second); err != nil {
		return ids, err
	}
	return ids, nil
}

// unpause unpauses a Kubernetes cluster
func unpause(cr cruntime.Manager, r command.Runner, namespaces []string) ([]string, error) {
	ids, err := cr.ListContainers(cruntime.ListContainersOptions{State: cruntime.Paused, Namespaces: namespaces})
	if err != nil {
		return ids, errors.Wrap(err, "list paused")
	}

	if len(ids) == 0 {
		klog.Warningf("no paused containers found")
	} else if err := cr.UnpauseContainers(ids); err != nil {
		return ids, errors.Wrap(err, "unpause")
	}

	sm := sysinit.New(r)

	if err := sm.Start("kubelet"); err != nil {
		return ids, errors.Wrap(err, "kubelet start")
	}

	return ids, nil
}

// CheckIfPaused checks if the Kubernetes cluster is paused
func CheckIfPaused(cr cruntime.Manager, namespaces []string) (bool, error) {
	ids, err := cr.ListContainers(cruntime.ListContainersOptions{State: cruntime.Paused, Namespaces: namespaces})
	if err != nil {
		return true, errors.Wrap(err, "list paused")
	}

	if len(ids) > 0 {
		return true, nil
	}

	return false, nil
}
