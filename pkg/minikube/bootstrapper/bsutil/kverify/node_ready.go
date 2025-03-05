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

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

// WaitNodeCondition waits for specified condition of node name.
func WaitNodeCondition(cs *kubernetes.Clientset, name string, condition core.NodeConditionType, timeout time.Duration) error {
	klog.Infof("waiting up to %v for node %q to be %q ...", timeout, name, condition)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s for node %q to be %q ...", time.Since(start), name, condition)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	lap := time.Now()
	checkCondition := func(ctx context.Context) (bool, error) {
		status, err := nodeConditionStatus(ctx, cs, name, condition)
		// done if node has condition
		if status == core.ConditionTrue {
			klog.Infof("node %q is %q", name, condition)
			return true, nil
		}
		// retry in all other cases, decrease log spam
		if time.Since(lap) > (2 * time.Second) {
			if err != nil {
				klog.Warningf("error getting node %q condition %q status (will retry): %v", name, condition, err)
			} else {
				klog.Warningf("node %q has %q:%q status (will retry)", name, condition, status)
			}
			lap = time.Now()
		}
		return false, nil
	}
	if err := wait.PollUntilContextCancel(ctx, kconst.APICallRetryInterval, true, checkCondition); err != nil {
		return fmt.Errorf("WaitNodeCondition: %w", err)
	}

	return nil
}

// nodeConditionStatus checks if node exists and returns condition status.
func nodeConditionStatus(ctx context.Context, cs *kubernetes.Clientset, name string, condition core.NodeConditionType) (core.ConditionStatus, error) {
	node, err := cs.CoreV1().Nodes().Get(ctx, name, meta.GetOptions{})
	if err != nil {
		return core.ConditionUnknown, err
	}
	// check if node has the condition
	for _, c := range node.Status.Conditions {
		if c.Type == condition {
			return c.Status, nil
		}
	}
	// assume transient error
	return core.ConditionUnknown, fmt.Errorf("node %q does not have %q condition type: %+v", name, condition, node.Status)
}
