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
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

// WaitExtra calls waitPodCondition for all system-critical pods including those with specified labels.
func WaitExtra(cs *kubernetes.Clientset, labels []string, timeout time.Duration) error {
	klog.Infof("extra waiting up to %v for all system-critical pods including labels %v to be %q ...", timeout, labels, core.PodReady)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s for extra waiting for all system-critical and pods with labels %v to be %q ...", time.Since(start), labels, core.PodReady)
	}()

	pods, err := cs.CoreV1().Pods(meta.NamespaceSystem).List(context.Background(), meta.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing pods in %q namespace: %w", meta.NamespaceSystem, err)
	}

	for _, pod := range pods.Items {
		if time.Since(start) > timeout {
			return fmt.Errorf("timed out waiting %v for all system-critical and pods with labels %v to be %q", timeout, labels, core.NodeReady)
		}

		for k, v := range pod.Labels {
			label := fmt.Sprintf("%s=%s", k, v)
			match := false
			for _, l := range labels {
				if l == label {
					match = true
					break
				}
			}
			// ignore system-critical pods' non-essential labels
			if !match && pod.Namespace != meta.NamespaceSystem && k != "k8s-app" && k != "component" {
				continue
			}
			if match || pod.Spec.PriorityClassName == "system-cluster-critical" || pod.Spec.PriorityClassName == "system-node-critical" {
				if err := waitPodCondition(cs, pod.Name, pod.Namespace, core.PodReady, timeout); err != nil {
					klog.Errorf("WaitExtra: %v", err)
				}
				break
			}
		}
	}

	return nil
}

// waitPodCondition waits for specified condition of podName in a namespace.
func waitPodCondition(cs *kubernetes.Clientset, name, namespace string, condition core.PodConditionType, timeout time.Duration) error {
	klog.Infof("waiting up to %v for pod %q in %q namespace to be %q ...", timeout, name, namespace, condition)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %v waiting for pod %q in %q namespace to be %q ...", time.Since(start), name, namespace, condition)
	}()

	lap := time.Now()
	checkCondition := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("timed out waiting %v for pod %q in %q namespace to be %q (will not retry!)", timeout, name, namespace, condition)
		}

		status, reason := podConditionStatus(cs, name, namespace, condition)
		if status == core.ConditionTrue {
			klog.Info(reason)
			return true, nil
		}
		// return immediately: status == core.ConditionUnknown
		if status == core.ConditionUnknown {
			klog.Info(reason)
			return false, fmt.Errorf(reason)
		}
		// reduce log spam
		if time.Since(lap) > (2 * time.Second) {
			klog.Info(reason)
			lap = time.Now()
		}
		// return immediately: status == core.ConditionFalse
		return false, nil
	}
	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, checkCondition); err != nil {
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

	// check if undelying node is Ready - in case we got stale data about the pod
	if pod.Spec.NodeName != "" {
		if status, reason := nodeConditionStatus(cs, pod.Spec.NodeName, core.NodeReady); status != core.ConditionTrue {
			return core.ConditionUnknown, fmt.Sprintf("node %q hosting pod %q in %q namespace is currently not %q (skipping!): %v", pod.Spec.NodeName, name, namespace, core.NodeReady, reason)
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
