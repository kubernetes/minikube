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
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

// WaitExtra calls WaitForPodReadyByLabel for each pod in labels list and returns any errors occurred.
func WaitExtra(cs *kubernetes.Clientset, labels []string, timeout time.Duration) error {
	klog.Infof("extra waiting for kube-system core pods %s to be Ready ...", labels)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s for extra waiting for kube-system core pods to be Ready ...", time.Since(start))
	}()

	var errs []string
	for _, label := range labels {
		if err := WaitForPodReadyByLabel(cs, label, "kube-system", timeout); err != nil {
			errs = append(errs, fmt.Sprintf("%q: %q", label, err.Error()))
		}
	}
	if errs != nil {
		return fmt.Errorf(strings.Join(errs, ", "))
	}

	return nil
}

// WaitForPodReadyByLabel waits for pod with label ([key:]val) in a namespace to be in Ready condition.
// If namespace is not provided, it defaults to "kube-system".
// If label key is not provided, it will try with "component" and "k8s-app".
func WaitForPodReadyByLabel(cs *kubernetes.Clientset, label, namespace string, timeout time.Duration) error {
	klog.Infof("waiting %v for pod with %q label in %q namespace to be Ready ...", timeout, label, namespace)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %v to run WaitForPodReadyByLabel for pod with %q label in %q namespace ...", time.Since(start), label, namespace)
	}()

	if namespace == "" {
		namespace = "kube-system"
	}

	lkey := ""
	lval := ""
	l := strings.Split(label, ":")
	switch len(l) {
	case 1: // treat as no label key provided, just val
		lval = strings.TrimSpace(l[0])
	case 2:
		lkey = strings.TrimSpace(l[0])
		lval = strings.TrimSpace(l[1])
	default:
		return fmt.Errorf("pod label %q is malformed", label)
	}

	lap := time.Now()
	checkReady := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("wait for pod with %q label in %q namespace to be Ready timed out", label, namespace)
		}
		pods, err := cs.CoreV1().Pods(namespace).List(meta.ListOptions{})
		if err != nil {
			klog.Infof("error listing pods in %q namespace, will retry: %v", namespace, err)
			return false, nil
		}
		for _, pod := range pods.Items {
			for k, v := range pod.ObjectMeta.Labels {
				if ((lkey == "" && (k == "component" || k == "k8s-app")) || lkey == k) && v == lval {
					ready, reason := IsPodReady(&pod)
					if ready {
						klog.Info(reason)
						return true, nil
					}
					// reduce log spam
					if time.Since(lap) > (1 * time.Second) {
						klog.Info(reason)
						lap = time.Now()
					}
					return false, nil
				}
			}
		}
		klog.Infof("pod with %q label in %q namespace was not found, will retry", label, namespace)
		return false, nil
	}
	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, checkReady); err != nil {
		return errors.Wrapf(err, "wait pod Ready")
	}

	return nil
}

// WaitForPodReadyByName waits for pod with name in a namespace to be in Ready condition.
// If namespace is not provided, it defaults to "kube-system".
func WaitForPodReadyByName(cs *kubernetes.Clientset, name, namespace string, timeout time.Duration) error {
	klog.Infof("waiting %v for pod %q in %q namespace to be Ready ...", timeout, name, namespace)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %v to run WaitForPodReadyByName for pod %q in %q namespace ...", time.Since(start), name, namespace)
	}()

	if namespace == "" {
		namespace = "kube-system"
	}

	lap := time.Now()
	checkReady := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("wait for pod %q in %q namespace to be Ready timed out", name, namespace)
		}
		pod, err := cs.CoreV1().Pods(namespace).Get(name, meta.GetOptions{})
		if err != nil {
			klog.Infof("error getting pod %q in %q namespace, will retry: %v", name, namespace, err)
			return false, nil
		}
		ready, reason := IsPodReady(pod)
		if ready {
			klog.Info(reason)
			return true, nil
		}
		// reduce log spam
		if time.Since(lap) > (1 * time.Second) {
			klog.Info(reason)
			lap = time.Now()
		}
		return false, nil
	}
	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, checkReady); err != nil {
		return errors.Wrapf(err, "wait pod Ready")
	}

	return nil
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
