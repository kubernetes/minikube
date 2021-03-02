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
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

// WaitNodeCondition waits for specified condition of node name.
func WaitNodeCondition(cs *kubernetes.Clientset, name string, condition core.NodeConditionType, timeout time.Duration) error {
	klog.Infof("waiting up to %v for node %q to be %q ...", timeout, name, condition)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %v waiting for node %q to be %q ...", time.Since(start), name, condition)
	}()

	lap := time.Now()
	checkCondition := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("timed out waiting %v for node %q to be %q (will not retry!)", timeout, name, condition)
		}

		status, reason := nodeConditionStatus(cs, name, condition)
		if status == core.ConditionTrue {
			klog.Info(reason)
			return true, nil
		}
		if status == core.ConditionUnknown {
			klog.Info(reason)
			return false, fmt.Errorf(reason)
		}
		// reduce log spam
		if time.Since(lap) > (2 * time.Second) {
			klog.Info(reason)
			lap = time.Now()
		}
		return false, nil
	}
	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, checkCondition); err != nil {
		return fmt.Errorf("waitNodeCondition: %w", err)
	}

	return nil
}

// nodeConditionStatus returns if node is in specified condition and verbose reason.
func nodeConditionStatus(cs *kubernetes.Clientset, name string, condition core.NodeConditionType) (status core.ConditionStatus, reason string) {
	node, err := cs.CoreV1().Nodes().Get(context.Background(), name, meta.GetOptions{})
	if err != nil {
		return core.ConditionUnknown, fmt.Sprintf("error getting node %q: %v", name, err)
	}

	for _, c := range node.Status.Conditions {
		if c.Type == condition {
			return c.Status, fmt.Sprintf("node %q has status %q:%q", node.Name, condition, c.Status)
		}
	}

	// assume transient condition
	return core.ConditionFalse, fmt.Sprintf("node %q doesn't have %q status: %+v", node.Name, condition, node.Status)
}
