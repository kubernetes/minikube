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

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

// WaitForNodeReady waits till kube client reports node status as "ready"
func WaitForNodeReady(cs *kubernetes.Clientset, timeout time.Duration) error {
	klog.Infof("waiting %s for node status to be ready ...", timeout)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s to wait for WaitForNodeReady...", time.Since(start))
	}()
	checkReady := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("wait for node to be ready timed out")
		}
		ns, err := cs.CoreV1().Nodes().List(meta.ListOptions{})
		if err != nil {
			klog.Infof("error listing nodes will retry: %v", err)
			return false, nil
		}

		for _, n := range ns.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == v1.NodeReady && c.Status != v1.ConditionTrue {
					klog.Infof("node %q has unwanted condition %q : Reason %q Message: %q. will try. ", n.Name, c.Type, c.Reason, c.Message)
					return false, nil
				}
			}
		}
		return true, nil
	}
	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, checkReady); err != nil {
		return errors.Wrapf(err, "wait node ready")
	}
	return nil
}
