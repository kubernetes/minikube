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

package kapi

import (
	"context"
	"fmt"
	"path"
	"time"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/minikube/vmpath"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

var (
	// ReasonableMutateTime is how long to wait for basic object mutations, such as deletions, to show up
	ReasonableMutateTime = time.Minute * 2
	// ReasonableStartTime is how long to wait for pods to start
	ReasonableStartTime = time.Minute * 5
)

// ClientConfig returns the client configuration for a kubectl context
func ClientConfig(ctx string) (*rest.Config, error) {
	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, &clientcmd.ConfigOverrides{CurrentContext: ctx})
	c, err := cc.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("client config: %v", err)
	}
	c = proxy.UpdateTransport(c)
	klog.V(1).Infof("client config for %s: %+v", ctx, c)
	return c, nil
}

// Client gets the Kubernetes client for a kubectl context name
func Client(ctx string) (*kubernetes.Clientset, error) {
	c, err := ClientConfig(ctx)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(c)
}

// WaitForPods waits for all matching pods to become Running or finish successfully and at least one matching pod exists.
func WaitForPods(c kubernetes.Interface, ns string, selector string, timeOut ...time.Duration) error {
	start := time.Now()
	klog.Infof("Waiting for pod with label %q in ns %q ...", selector, ns)
	lastKnownPodNumber := -1
	f := func(ctx context.Context) (bool, error) {
		listOpts := meta.ListOptions{LabelSelector: selector}
		pods, err := c.CoreV1().Pods(ns).List(ctx, listOpts)
		if err != nil {
			klog.Infof("temporary error: getting Pods with label selector %q : [%v]\n", selector, err)
			return false, nil
		}

		if lastKnownPodNumber != len(pods.Items) {
			klog.Infof("Found %d Pods for label selector %s\n", len(pods.Items), selector)
			lastKnownPodNumber = len(pods.Items)
		}

		if len(pods.Items) == 0 {
			return false, nil
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != core.PodRunning && pod.Status.Phase != core.PodSucceeded {
				klog.Infof("waiting for pod %q, current state: %s: [%v]\n", selector, pod.Status.Phase, err)
				return false, nil
			}
		}
		return true, nil
	}
	t := ReasonableStartTime
	if timeOut != nil {
		t = timeOut[0]
	}
	err := wait.PollUntilContextTimeout(context.Background(), kconst.APICallRetryInterval, t, true, f)
	klog.Infof("duration metric: took %s to wait for %s ...", time.Since(start), selector)
	return err
}

// WaitForDeploymentToStabilize waits till the Deployment has a matching generation/replica count between spec and status. used by integration tests
func WaitForDeploymentToStabilize(c kubernetes.Interface, ns, name string, timeout time.Duration) error {
	options := meta.ListOptions{FieldSelector: fields.Set{
		"metadata.name":      name,
		"metadata.namespace": ns,
	}.AsSelector().String()}

	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), timeout)
	defer cancel()

	w, err := c.AppsV1().Deployments(ns).Watch(ctx, options)
	if err != nil {
		return err
	}
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
			klog.Infof("Waiting for deployment %s to stabilize, generation %v observed generation %v spec.replicas %d status.replicas %d",
				name, dp.Generation, dp.Status.ObservedGeneration, *(dp.Spec.Replicas), dp.Status.Replicas)
		}
		return false, nil
	})
	return err
}

// WaitForService waits until the service appears (exist == true), or disappears (exist == false)
func WaitForService(c kubernetes.Interface, namespace, name string, exist bool, interval, timeout time.Duration) error {
	err := wait.PollUntilContextTimeout(context.Background(), interval, timeout, true, func(ctx context.Context) (bool, error) {
		_, err := c.CoreV1().Services(namespace).Get(ctx, name, meta.GetOptions{})
		switch {
		case err == nil:
			klog.Infof("Service %s in namespace %s found.", name, namespace)
			return exist, nil
		case apierr.IsNotFound(err):
			klog.Infof("Service %s in namespace %s disappeared.", name, namespace)
			return !exist, nil
		case !IsRetryableAPIError(err):
			klog.Info("Non-retryable failure while getting service.")
			return false, err
		default:
			klog.Infof("Get service %s in namespace %s failed: %v", name, namespace, err)
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

// KubectlBinaryPath returns the path to kubectl on the node
func KubectlBinaryPath(version string) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", version, "kubectl")
}

// ScaleDeployment tries to set the number of deployment replicas in namespace and context.
// It will retry (usually needed due to "the object has been modified; please apply your changes to the latest version and try again" error) up to ReasonableMutateTime to ensure target scale is achieved.
func ScaleDeployment(kcontext, namespace, deploymentName string, replicas int) error {
	client, err := Client(kcontext)
	if err != nil {
		return fmt.Errorf("client: %v", err)
	}

	err = wait.PollUntilContextTimeout(context.Background(), kconst.APICallRetryInterval, ReasonableMutateTime, true, func(ctx context.Context) (bool, error) {
		scale, err := client.AppsV1().Deployments(namespace).GetScale(ctx, deploymentName, meta.GetOptions{})
		if err != nil {
			if !IsRetryableAPIError(err) {
				return false, fmt.Errorf("non-retryable failure while getting %q deployment scale: %v", deploymentName, err)
			}
			klog.Warningf("failed getting %q deployment scale, will retry: %v", deploymentName, err)
			return false, nil
		}
		if scale.Spec.Replicas != int32(replicas) {
			scale.Spec.Replicas = int32(replicas)
			if _, err = client.AppsV1().Deployments(namespace).UpdateScale(ctx, deploymentName, scale, meta.UpdateOptions{}); err != nil {
				if !IsRetryableAPIError(err) {
					return false, fmt.Errorf("non-retryable failure while rescaling %s deployment: %v", deploymentName, err)
				}
				klog.Warningf("failed rescaling %s deployment, will retry: %v", deploymentName, err)
			}
			// repeat (if change was successful - once again to check & confirm requested scale)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		klog.Warningf("failed rescaling %q deployment in %q namespace and %q context to %d replicas: %v", deploymentName, namespace, kcontext, replicas, err)
		return err
	}
	klog.Infof("%q deployment in %q namespace and %q context rescaled to %d replicas", deploymentName, namespace, kcontext, replicas)

	return nil
}
