/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

// WaitExtra calls waitPodCondition for all (at least one) kube-system pods having one of specified labels to be "Ready" on profile cluster.
func WaitExtra(profile string, labels []string, timeout time.Duration) error {
	klog.Infof("extra waiting up to %v for all %q pods having one of %v labels to be %q ...", timeout, meta.NamespaceSystem, labels, core.PodReady)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s for extra waiting for all %q pods having one of %v labels to be %q ...", time.Since(start), meta.NamespaceSystem, labels, core.PodReady)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cs, err := kapi.Client(profile)
	if err != nil {
		return fmt.Errorf("failed to get kube client: %v", err)
	}

	// podsReady poll function checks if all (at least one) pods in the namespace having the label is Ready
	var label string
	podsReady := func(ctx context.Context) (bool, error) {
		pods, err := cs.CoreV1().Pods(meta.NamespaceSystem).List(ctx, meta.ListOptions{LabelSelector: label})
		if err != nil {
			klog.Warningf("error listing pods in %q namespace with %q label, will retry: %v", meta.NamespaceSystem, label, err)
			return false, nil
		}
		if len(pods.Items) == 0 {
			klog.Warningf("no pods in %q namespace with %q label found, will retry", meta.NamespaceSystem, label)
			return false, nil
		}
		for _, pod := range pods.Items {
			if err := waitPodCondition(ctx, cs, pod.Name, pod.Namespace, core.PodReady); err != nil {
				klog.Warningf("not all pods in %q namespace with %q label are %q, will retry: %v", meta.NamespaceSystem, label, core.PodReady, err)
				return false, nil
			}
		}
		return true, nil
	}
	for _, l := range labels {
		label = l
		if err := wait.PollUntilContextCancel(ctx, kconst.APICallRetryInterval, true, podsReady); err != nil {
			return fmt.Errorf("WaitExtra: %w", err)
		}
	}

	return nil
}

// waitPodCondition waits for specified condition of pod name in namespace.
func waitPodCondition(ctx context.Context, cs *kubernetes.Clientset, name, namespace string, condition core.PodConditionType) error {
	klog.Infof("waiting for pod %q in %q namespace to be %q or be gone ...", name, namespace, condition)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s for pod %q in %q namespace to be %q or be gone ...", time.Since(start), name, namespace, condition)
	}()

	lap := time.Now()
	checkCondition := func(ctx context.Context) (bool, error) {
		status, err := podConditionStatus(ctx, cs, name, namespace, condition)
		// done if pod has condition
		if status == core.ConditionTrue {
			klog.Infof("pod %q is %q", name, condition)
			return true, nil
		}
		// back off if pod or node is gone
		if kerrors.IsNotFound(err) || status == core.TaintNodeUnreachable {
			klog.Infof("pod %q in %q namespace is gone: %v", name, namespace, err)
			return true, nil
		}
		// retry in all other cases, decrease log spam
		if time.Since(lap) > (2 * time.Second) {
			klog.Warningf("pod %q is not %q, error: %v", name, condition, err)
			lap = time.Now()
		}
		return false, nil
	}
	if err := wait.PollUntilContextCancel(ctx, kconst.APICallRetryInterval, true, checkCondition); err != nil {
		return fmt.Errorf("waitPodCondition: %w", err)
	}

	return nil
}

// podConditionStatus returns if pod is in specified condition.
func podConditionStatus(ctx context.Context, cs *kubernetes.Clientset, name, namespace string, condition core.PodConditionType) (core.ConditionStatus, error) {
	pod, err := cs.CoreV1().Pods(namespace).Get(ctx, name, meta.GetOptions{})
	if err != nil {
		return core.ConditionUnknown, fmt.Errorf("getting pod %q in %q namespace (will retry): %w", name, namespace, err)
	}
	// check if pod is scheduled on any node
	if pod.Spec.NodeName == "" {
		return core.ConditionUnknown, fmt.Errorf("pod %q in %q namespace is not scheduled on any node (will retry): %+v", name, namespace, pod.Status)
	}
	// check if node exists and is Ready (KubeAPI)
	nodeReadyStatus, err := nodeConditionStatus(ctx, cs, pod.Spec.NodeName, core.NodeReady)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return core.TaintNodeUnreachable, fmt.Errorf("node %q hosting pod %q is not found/running (skipping!): %v", pod.Spec.NodeName, name, err)
		}
		return core.ConditionUnknown, fmt.Errorf("node %q hosting pod %q is not %q (will retry): %v", pod.Spec.NodeName, name, core.NodeReady, err)
	}
	if nodeReadyStatus != core.ConditionTrue {
		return core.ConditionUnknown, fmt.Errorf("node %q hosting pod %q is not %q (will retry)", pod.Spec.NodeName, name, core.NodeReady)
	}
	// check if pod has the condition
	for _, c := range pod.Status.Conditions {
		if c.Type == condition {
			return c.Status, nil
		}
	}
	// assume transient error
	return core.ConditionUnknown, fmt.Errorf("pod %q does not have %q condition type: %+v", name, condition, pod.Status)
}

// IsPodReady returns if pod is Ready.
func IsPodReady(pod *core.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == core.PodReady {
			return c.Status == core.ConditionTrue
		}
	}
	// assume transient error
	return false
}
