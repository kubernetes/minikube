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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

var (
	// ReasonableMutateTime is how long to wait for basic object mutations, such as deletions, to show up
	ReasonableMutateTime = time.Minute * 1
	// ReasonableStartTime is how long to wait for pods to start, considering dependency chains & slow networks.
	ReasonableStartTime = time.Minute * 10
)

// PodStore stores pods
type PodStore struct {
	cache.Store
	stopCh    chan struct{}
	Reflector *cache.Reflector
}

// List lists the pods
func (s *PodStore) List() []*v1.Pod {
	objects := s.Store.List()
	pods := make([]*v1.Pod, 0)
	for _, o := range objects {
		pods = append(pods, o.(*v1.Pod))
	}
	return pods
}

// Stop stops the pods
func (s *PodStore) Stop() {
	close(s.stopCh)
}

// GetClient gets the client from config
func GetClient() (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Error creating kubeConfig: %v", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new client from kubeConfig.ClientConfig()")
	}
	return client, nil
}

// NewPodStore creates a new PodStore
func NewPodStore(c kubernetes.Interface, namespace string, label labels.Selector, field fields.Selector) *PodStore {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.LabelSelector = label.String()
			options.FieldSelector = field.String()
			obj, err := c.Core().Pods(namespace).List(options)
			return runtime.Object(obj), err
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = label.String()
			options.FieldSelector = field.String()
			return c.Core().Pods(namespace).Watch(options)
		},
	}
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)
	stopCh := make(chan struct{})
	reflector := cache.NewReflector(lw, &v1.Pod{}, store, 0)
	go reflector.Run(stopCh)
	return &PodStore{Store: store, stopCh: stopCh, Reflector: reflector}
}

// StartPods starts all pods
func StartPods(c kubernetes.Interface, namespace string, pod v1.Pod, waitForRunning bool) error {
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
		listOpts := metav1.ListOptions{LabelSelector: label.String()}
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
			if pod.Status.Phase != v1.PodRunning {
				return false, nil
			}
		}

		return true, nil
	})
}

// WaitForPodDelete waits for a pod to be deleted
func WaitForPodDelete(c kubernetes.Interface, ns string, label labels.Selector) error {
	return wait.PollImmediate(constants.APICallRetryInterval, ReasonableMutateTime, func() (bool, error) {
		listOpts := metav1.ListOptions{LabelSelector: label.String()}
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
		events, err := c.Events().Events("default").List(metav1.ListOptions{})
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

// WaitForRCToStabilize waits till the RC has a matching generation/replica count between spec and status.
func WaitForRCToStabilize(c kubernetes.Interface, ns, name string, timeout time.Duration) error {
	options := metav1.ListOptions{FieldSelector: fields.Set{
		"metadata.name":      name,
		"metadata.namespace": ns,
	}.AsSelector().String()}
	w, err := c.Core().ReplicationControllers(ns).Watch(options)
	if err != nil {
		return err
	}
	_, err = watch.Until(timeout, w, func(event watch.Event) (bool, error) {
		switch event.Type {
		case watch.Deleted:
			return false, apierrs.NewNotFound(schema.GroupResource{Resource: "replicationcontrollers"}, "")
		}
		switch rc := event.Object.(type) {
		case *v1.ReplicationController:
			if rc.Name == name && rc.Namespace == ns &&
				rc.Generation <= rc.Status.ObservedGeneration &&
				*(rc.Spec.Replicas) == rc.Status.Replicas {
				return true, nil
			}
			glog.Infof("Waiting for rc %s to stabilize, generation %v observed generation %v spec.replicas %d status.replicas %d",
				name, rc.Generation, rc.Status.ObservedGeneration, *(rc.Spec.Replicas), rc.Status.Replicas)
		}
		return false, nil
	})
	return err
}

// WaitForDeploymentToStabilize waits till the Deployment has a matching generation/replica count between spec and status.
func WaitForDeploymentToStabilize(c kubernetes.Interface, ns, name string, timeout time.Duration) error {
	options := metav1.ListOptions{FieldSelector: fields.Set{
		"metadata.name":      name,
		"metadata.namespace": ns,
	}.AsSelector().String()}
	w, err := c.AppsV1().Deployments(ns).Watch(options)
	if err != nil {
		return err
	}
	_, err = watch.Until(timeout, w, func(event watch.Event) (bool, error) {
		switch event.Type {
		case watch.Deleted:
			return false, apierrs.NewNotFound(schema.GroupResource{Resource: "deployments"}, "")
		}
		switch dp := event.Object.(type) {
		case *appsv1.Deployment:
			if dp.Name == name && dp.Namespace == ns &&
				dp.Generation <= dp.Status.ObservedGeneration &&
				*(dp.Spec.Replicas) == dp.Status.Replicas {
				return true, nil
			}
			glog.Infof("Waiting for deployment %s to stabilize, generation %v observed generation %v spec.replicas %d status.replicas %d",
				name, dp.Generation, dp.Status.ObservedGeneration, *(dp.Spec.Replicas), dp.Status.Replicas)
		}
		return false, nil
	})
	return err
}

// WaitForService waits until the service appears (exist == true), or disappears (exist == false)
func WaitForService(c kubernetes.Interface, namespace, name string, exist bool, interval, timeout time.Duration) error {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		_, err := c.Core().Services(namespace).Get(name, metav1.GetOptions{})
		switch {
		case err == nil:
			glog.Infof("Service %s in namespace %s found.", name, namespace)
			return exist, nil
		case apierrs.IsNotFound(err):
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
		list, err := c.Core().Endpoints(namespace).List(metav1.ListOptions{})
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

func countEndpointsNum(e *v1.Endpoints) int {
	num := 0
	for _, sub := range e.Subsets {
		num += len(sub.Addresses)
	}
	return num
}

// IsRetryableAPIError returns if this error is retryable or not
func IsRetryableAPIError(err error) bool {
	return apierrs.IsTimeout(err) || apierrs.IsServerTimeout(err) || apierrs.IsTooManyRequests(err) || apierrs.IsInternalError(err)
}
