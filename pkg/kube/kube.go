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

package kube

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/minikube/proxy"
)

var (
	// ReasonableMutateTime is how long to wait for basic object mutations, such as deletions, to show up
	ReasonableMutateTime = time.Minute * 2
	// ReasonableStartTime is how long to wait for pods to start
	ReasonableStartTime = time.Minute * 6
)

// Client gets the kuberentes client from default kubeconfig
func Client(kubectlContext ...string) (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if kubectlContext != nil {
		configOverrides = &clientcmd.ConfigOverrides{
			CurrentContext: kubectlContext[0],
		}
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating kubeConfig: %v", err)
	}
	config = proxy.UpdateTransport(config)
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "error creating new client from kubeConfig.ClientConfig()")
	}
	return client, nil
}

// WaitForPodsWithLabelRunning waits for all matching pods to become Running and at least one matching pod exists.
func WaitForPodsWithLabelRunning(c kubernetes.Interface, ns string, label labels.Selector) error {
	glog.Infof("Waiting for pod with label %q in ns %q ...", ns, label)
	lastKnownPodNumber := -1
	return wait.PollImmediate(constants.APICallRetryInterval, ReasonableStartTime, func() (bool, error) {
		listOpts := meta.ListOptions{LabelSelector: label.String()}
		pods, err := c.CoreV1().Pods(ns).List(listOpts)
		if err != nil {
			glog.Errorf("error getting Pods with label selector %q [%v]\n", label.String(), err)
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

// WaitForRCToStabilize waits till the RC has a matching generation/replica count between spec and status.
func WaitForRCToStabilize(c kubernetes.Interface, ns, name string, timeout time.Duration) error {
	options := meta.ListOptions{FieldSelector: fields.Set{
		"metadata.name":      name,
		"metadata.namespace": ns,
	}.AsSelector().String()}
	w, err := c.CoreV1().ReplicationControllers(ns).Watch(options)
	if err != nil {
		return err
	}

	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), timeout)
	defer cancel()

	_, err = watchtools.UntilWithoutRetry(ctx, w, func(event watch.Event) (bool, error) {
		if event.Type == watch.Deleted {
			return false, apierr.NewNotFound(schema.GroupResource{Resource: "replicationcontrollers"}, "")
		}

		rc, ok := event.Object.(*core.ReplicationController)
		if ok {
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
	options := meta.ListOptions{FieldSelector: fields.Set{
		"metadata.name":      name,
		"metadata.namespace": ns,
	}.AsSelector().String()}
	w, err := c.AppsV1().Deployments(ns).Watch(options)
	if err != nil {
		return err
	}

	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), timeout)
	defer cancel()

	_, err = watchtools.UntilWithoutRetry(ctx, w, func(event watch.Event) (bool, error) {
		if event.Type == watch.Deleted {
			return false, apierr.NewNotFound(schema.GroupResource{Resource: "deployments"}, "")
		}
		dp, ok := event.Object.(*apps.Deployment)
		if ok {
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

// IsRetryableAPIError returns if this error is retryable or not
func IsRetryableAPIError(err error) bool {
	return apierr.IsTimeout(err) || apierr.IsServerTimeout(err) || apierr.IsTooManyRequests(err) || apierr.IsInternalError(err)
}
