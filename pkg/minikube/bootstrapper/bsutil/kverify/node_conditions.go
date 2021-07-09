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
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/util/retry"
)

// NodeCondition represents a favorable or unfavorable node condition.
type NodeCondition struct {
	Type    v1.NodeConditionType
	Status  v1.ConditionStatus
	Reason  string
	Message string
}

// DiskPressure detects if the condition is disk pressure
func (pc *NodeCondition) DiskPressure() bool {
	return pc.Type == v1.NodeDiskPressure && pc.Status == v1.ConditionTrue
}

// MemoryPressure detects if the condition is memory pressure
func (pc *NodeCondition) MemoryPressure() bool {
	return pc.Type == v1.NodeMemoryPressure && pc.Status == v1.ConditionTrue
}

// PIDPressure detects if the condition is PID pressure
func (pc *NodeCondition) PIDPressure() bool {
	return pc.Type == v1.NodePIDPressure && pc.Status == v1.ConditionTrue
}

// NetworkUnavailable detects if the condition is PID pressure
func (pc *NodeCondition) NetworkUnavailable() bool {
	return pc.Type == v1.NodeNetworkUnavailable && pc.Status == v1.ConditionTrue
}

const errTextFormat = "node has unwanted condition %q : Reason %q Message: %q"

// ErrMemoryPressure is thrown when there is node memory pressure condition
type ErrMemoryPressure struct {
	NodeCondition
}

func (e *ErrMemoryPressure) Error() string {
	return fmt.Sprintf(errTextFormat, e.Type, e.Reason, e.Message)
}

// ErrDiskPressure is thrown when there is node disk pressure condition
type ErrDiskPressure struct {
	NodeCondition
}

func (e *ErrDiskPressure) Error() string {
	return fmt.Sprintf(errTextFormat, e.Type, e.Reason, e.Message)
}

// ErrPIDPressure is thrown when there is node PID pressure condition
type ErrPIDPressure struct {
	NodeCondition
}

func (e *ErrPIDPressure) Error() string {
	return fmt.Sprintf(errTextFormat, e.Type, e.Reason, e.Message)
}

// ErrNetworkNotReady is thrown when there is node condition is network not ready
type ErrNetworkNotReady struct {
	NodeCondition
}

func (e *ErrNetworkNotReady) Error() string {
	return fmt.Sprintf(errTextFormat, e.Type, e.Reason, e.Message)
}

// NodePressure verfies that node is not under disk, memory, pid or network pressure.
func NodePressure(cs *kubernetes.Clientset) error {
	klog.Info("verifying NodePressure condition ...")
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s to run NodePressure ...", time.Since(start))
	}()

	var ns *v1.NodeList
	var err error

	listNodes := func() error {
		ns, err = cs.CoreV1().Nodes().List(context.Background(), meta.ListOptions{})
		return err
	}

	err = retry.Expo(listNodes, kconst.APICallRetryInterval, 2*time.Minute)
	if err != nil {
		return errors.Wrap(err, "list nodes retry")
	}

	for _, n := range ns.Items {
		klog.Infof("node storage ephemeral capacity is %s", n.Status.Capacity.StorageEphemeral())
		klog.Infof("node cpu capacity is %s", n.Status.Capacity.Cpu().AsDec())
		for _, c := range n.Status.Conditions {
			pc := NodeCondition{Type: c.Type, Status: c.Status, Reason: c.Reason, Message: c.Message}
			if pc.DiskPressure() {
				return &ErrDiskPressure{
					NodeCondition: pc,
				}
			}

			if pc.MemoryPressure() {
				return &ErrMemoryPressure{
					NodeCondition: pc,
				}
			}

			if pc.PIDPressure() {
				return &ErrPIDPressure{
					NodeCondition: pc,
				}
			}

			if pc.NetworkUnavailable() {
				return &ErrNetworkNotReady{
					NodeCondition: pc,
				}
			}

		}
	}
	return nil
}
