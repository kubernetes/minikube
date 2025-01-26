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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

// WaitExtra calls waitPodCondition for all (at least one) kube-system pods having one of specified labels to be "Ready".
func WaitExtra(cs *kubernetes.Clientset, labels []string, timeout time.Duration) error {
	klog.Infof("extra waiting up to %v for all %q pods having one of %v labels to be %q ...", timeout, meta.NamespaceSystem, labels, core.PodReady)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s for extra waiting for all %q pods having one of %v labels to be %q ...", time.Since(start), meta.NamespaceSystem, labels, core.PodReady)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

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

// waitPodCondition waits for specified condition of podName in a namespace.
func waitPodCondition(ctx context.Context, cs *kubernetes.Clientset, name, namespace string, condition core.PodConditionType) error {
	klog.Infof("waiting for pod %q in %q namespace to be %q or be gone ...", name, namespace, condition)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s for pod %q in %q namespace to be %q or be gone ...", time.Since(start), name, namespace, condition)
	}()

	lap := time.Now()
	checkCondition := func(_ context.Context) (bool, error) {
		status, reason := podConditionStatus(cs, name, namespace, condition)
		// done if pod is ready
		if status == core.ConditionTrue {
			klog.Info(reason)
			return true, nil
		}

		// back off if pod condition is unknown or node is not ready - we check node healt elsewhere
		if status == core.ConditionUnknown || status == core.TaintNodeNotReady {
			klog.Warning(reason)
			return true, nil
		}

		// retry in all other cases (eg, node not ready, pod pending, pod not ready, etc.)
		// decrease log spam
		if time.Since(lap) > (2 * time.Second) {
			klog.Warning(reason)
			lap = time.Now()
		}
		return false, nil
	}
	if err := wait.PollUntilContextCancel(ctx, kconst.APICallRetryInterval, false, checkCondition); err != nil {
		return fmt.Errorf("waitPodCondition: %w", err)
	}

	return nil
}

// podConditionStatus returns if pod is in specified condition and verbose reason.
func podConditionStatus(cs *kubernetes.Clientset, name, namespace string, condition core.PodConditionType) (status core.ConditionStatus, reason string) {
	pod, err := cs.CoreV1().Pods(namespace).Get(context.Background(), name, meta.GetOptions{})
	if err != nil {
		return core.ConditionUnknown, fmt.Sprintf("error getting pod %q in %q namespace (skipping!): %v", name, namespace, err)
	}

	// check if undelying node is Ready - skip in case we got stale data about the pod
	if pod.Spec.NodeName != "" {
		if status, reason := nodeConditionStatus(cs, pod.Spec.NodeName, core.NodeReady); status != core.ConditionTrue {
			return core.TaintNodeNotReady, fmt.Sprintf("node %q hosting pod %q in %q namespace is not %q (skipping!): %v", pod.Spec.NodeName, name, namespace, core.NodeReady, reason)
		}
	}

	if pod.Status.Phase != core.PodRunning && pod.Status.Phase != core.PodPending {
		return core.ConditionUnknown, fmt.Sprintf("pod %q in %q namespace has status phase %q (skipping!): %+v", pod.Name, pod.Namespace, pod.Status.Phase, pod.Status)
	}

	for _, c := range pod.Status.Conditions {
		if c.Type == condition {
			return c.Status, fmt.Sprintf("pod %q in %q namespace has status %q:%q", pod.Name, pod.Namespace, condition, c.Status)
		}
	}

	// assume transient condition
	return core.ConditionFalse, fmt.Sprintf("pod %q in %q namespace doesn't have %q status: %+v", pod.Name, pod.Namespace, core.PodReady, pod.Status)
}

// IsPodReady returns if pod is Ready and verbose reason.
func IsPodReady(pod *core.Pod) (ready bool, reason string) {
	if pod.Status.Phase != core.PodRunning {
		return false, fmt.Sprintf("pod %q in %q namespace is not Running: %+v", pod.Name, pod.Namespace, pod.Status)
	}
	for _, c := range pod.Status.Conditions {
		if c.Type == core.PodReady {
			if c.Status != core.ConditionTrue {
				return false, fmt.Sprintf("pod %q in %q namespace is not Ready: %+v", pod.Name, pod.Namespace, c)
			}
			return true, fmt.Sprintf("pod %q in %q namespace is Ready: %+v", pod.Name, pod.Namespace, c)
		}
	}
	return false, fmt.Sprintf("pod %q in %q namespace does not have %q status: %+v", pod.Name, pod.Namespace, core.PodReady, pod.Status)
}
