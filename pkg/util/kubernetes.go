/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package util

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/minikube/proxy"
)

var (
	// ReasonableMutateTime is how long to wait for basic object mutations, such as deletions, to show up
	ReasonableMutateTime = time.Minute * 2
	// ReasonableStartTime is how long to wait for pods to start, considering dependency chains & slow networks.
	ReasonableStartTime = time.Minute * 10
)

// GetClient gets the client from config
func GetClient() (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Error creating kubeConfig: %v", err)
	}
	config = proxy.UpdateTransport(config)
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new client from kubeConfig.ClientConfig()")
	}
	return client, nil
}

// StartPods starts all pods
func StartPods(c kubernetes.Interface, namespace string, pod core.Pod, waitForRunning bool) error {
	pod.ObjectMeta.Labels["name"] = pod.Name
	if waitForRunning {
		label := labels.SelectorFromSet(labels.Set(map[string]string{"name": pod.Name}))
		err := WaitForPodsWithLabelRunning(c, namespace, label)
		if err != nil {
			return fmt.Errorf("Error waiting for pod %s to be running: %v", pod.Name, err)
		}
	}
	return nil
}

// WaitForPodsWithLabelRunning waits for all matching pods to become Running and at least one matching pod exists.
func WaitForPodsWithLabelRunning(c kubernetes.Interface, ns string, label labels.Selector) error {
	glog.Infof("Waiting for pod with label %q in ns %q ...", ns, label)
	lastKnownPodNumber := -1
	return wait.PollImmediate(constants.APICallRetryInterval, ReasonableStartTime, func() (bool, error) {
		listOpts := meta.ListOptions{LabelSelector: label.String()}
		pods, err := c.CoreV1().Pods(ns).List(listOpts)
		if err != nil {
			glog.Infof("error getting Pods with label selector %q [%v]\n", label.String(), err)
			return false, nil
		}

		if lastKnownPodNumber != len(pods.Items) {
			glog.Infof("Found %d Pods for label selector %s\n", len(pods.Items), label.String())
			lastKnownPodNumber = len(pods.Items)
		}

		if len(pods.Items) == 0 {
			return false, nil
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != core.PodRunning {
				return false, nil
			}
		}

		return true, nil
	})
}

// WaitForPodDelete waits for a pod to be deleted
func WaitForPodDelete(c kubernetes.Interface, ns string, label fmt.Stringer) error {
	return wait.PollImmediate(constants.APICallRetryInterval, ReasonableMutateTime, func() (bool, error) {
		listOpts := meta.ListOptions{LabelSelector: label.String()}
		pods, err := c.CoreV1().Pods(ns).List(listOpts)
		if err != nil {
			glog.Infof("error getting Pods with label selector %q [%v]\n", label.String(), err)
			return false, nil
		}
		return len(pods.Items) == 0, nil
	})
}

// WaitForEvent waits for the given event to appear
func WaitForEvent(c kubernetes.Interface, ns string, reason string) error {
	return wait.PollImmediate(constants.APICallRetryInterval, ReasonableMutateTime, func() (bool, error) {
		events, err := c.EventsV1beta1().Events("default").List(meta.ListOptions{})
		if err != nil {
			glog.Infof("error getting events: %v", err)
			return false, nil
		}
		for _, e := range events.Items {
			if e.Reason == reason {
				return true, nil
			}
		}
		return false, nil
	})
}

// WaitForService waits until the service appears (exist == true), or disappears (exist == false)
func WaitForService(c kubernetes.Interface, namespace, name string, exist bool, interval, timeout time.Duration) error {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		_, err := c.CoreV1().Services(namespace).Get(name, meta.GetOptions{})
		switch {
		case err == nil:
			glog.Infof("Service %s in namespace %s found.", name, namespace)
			return exist, nil
		case apierr.IsNotFound(err):
			glog.Infof("Service %s in namespace %s disappeared.", name, namespace)
			return !exist, nil
		case !IsRetryableAPIError(err):
			glog.Info("Non-retryable failure while getting service.")
			return false, err
		default:
			glog.Infof("Get service %s in namespace %s failed: %v", name, namespace, err)
			return false, nil
		}
	})
	if err != nil {
		stateMsg := map[bool]string{true: "to appear", false: "to disappear"}
		return fmt.Errorf("error waiting for service %s/%s %s: %v", namespace, name, stateMsg[exist], err)
	}
	return nil
}

//WaitForServiceEndpointsNum waits until the amount of endpoints that implement service to expectNum.
func WaitForServiceEndpointsNum(c kubernetes.Interface, namespace, serviceName string, expectNum int, interval, timeout time.Duration) error {
	return wait.Poll(interval, timeout, func() (bool, error) {
		glog.Infof("Waiting for amount of service:%s endpoints to be %d", serviceName, expectNum)
		list, err := c.CoreV1().Endpoints(namespace).List(meta.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, e := range list.Items {
			if e.Name == serviceName && countEndpointsNum(&e) == expectNum {
				return true, nil
			}
		}
		return false, nil
	})
}

func countEndpointsNum(e *core.Endpoints) int {
	num := 0
	for _, sub := range e.Subsets {
		num += len(sub.Addresses)
	}
	return num
}

// IsRetryableAPIError returns if this error is retryable or not
func IsRetryableAPIError(err error) bool {
	return apierr.IsTimeout(err) || apierr.IsServerTimeout(err) || apierr.IsTooManyRequests(err) || apierr.IsInternalError(err)
}
