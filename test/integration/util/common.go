/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"testing"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	commonutil "k8s.io/minikube/pkg/util"
)

// WaitForBusyboxRunning waits until busybox pod to be running
func WaitForBusyboxRunning(t *testing.T, namespace string, miniProfile string) error {
	client, err := commonutil.GetClient(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"integration-test": "busybox"}))
	return commonutil.WaitForPodsWithLabelRunning(client, namespace, selector)
}

// WaitForIngressControllerRunning waits until ingress controller pod to be running
func WaitForIngressControllerRunning(t *testing.T, miniProfile string) error {
	client, err := commonutil.GetClient(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	if err := commonutil.WaitForDeploymentToStabilize(client, "kube-system", "nginx-ingress-controller", time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for ingress-controller deployment to stabilize")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"app.kubernetes.io/name": "nginx-ingress-controller"}))
	if err := commonutil.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "waiting for ingress-controller pods")
	}

	return nil
}

// WaitForIngressDefaultBackendRunning waits until ingress default backend pod to be running
func WaitForIngressDefaultBackendRunning(t *testing.T, miniProfile string) error {
	client, err := commonutil.GetClient(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	if err := commonutil.WaitForDeploymentToStabilize(client, "kube-system", "default-http-backend", time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for default-http-backend deployment to stabilize")
	}

	if err := commonutil.WaitForService(client, "kube-system", "default-http-backend", true, time.Millisecond*500, time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for default-http-backend service to be up")
	}

	if err := commonutil.WaitForServiceEndpointsNum(client, "kube-system", "default-http-backend", 1, time.Second*3, time.Minute*10); err != nil {
		return errors.Wrap(err, "waiting for one default-http-backend endpoint to be up")
	}

	return nil
}

// WaitForGvisorControllerRunning waits for the gvisor controller pod to be running
func WaitForGvisorControllerRunning(t *testing.T, miniProfile string) error {
	client, err := commonutil.GetClient(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"kubernetes.io/minikube-addons": "gvisor"}))
	if err := commonutil.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "waiting for gvisor controller pod to stabilize")
	}
	return nil
}

// WaitForGvisorControllerDeleted waits for the gvisor controller pod to be deleted
func WaitForGvisorControllerDeleted(miniProfile string) error {
	client, err := commonutil.GetClient(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"kubernetes.io/minikube-addons": "gvisor"}))
	if err := commonutil.WaitForPodDelete(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "waiting for gvisor controller pod deletion")
	}
	return nil
}

// WaitForUntrustedNginxRunning waits for the untrusted nginx pod to start running
func WaitForUntrustedNginxRunning(miniProfile string) error {
	client, err := commonutil.GetClient(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx"}))
	if err := commonutil.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		return errors.Wrap(err, "waiting for nginx pods")
	}
	return nil
}

// WaitForFailedCreatePodSandBoxEvent waits for a FailedCreatePodSandBox event to appear
func WaitForFailedCreatePodSandBoxEvent(miniProfile string) error {
	client, err := commonutil.GetClient(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}
	if err := commonutil.WaitForEvent(client, "default", "FailedCreatePodSandBox"); err != nil {
		return errors.Wrap(err, "waiting for FailedCreatePodSandBox event")
	}
	return nil
}

// WaitForNginxRunning waits for nginx service to be up
func WaitForNginxRunning(t *testing.T, miniProfile string) error {
	client, err := commonutil.GetClient(miniProfile)

	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx"}))
	if err := commonutil.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		return errors.Wrap(err, "waiting for nginx pods")
	}

	if err := commonutil.WaitForService(client, "default", "nginx", true, time.Millisecond*500, time.Minute*10); err != nil {
		t.Errorf("Error waiting for nginx service to be up")
	}
	return nil
}

// Retry tries the callback for a number of attempts, with a delay between attempts
func Retry(t *testing.T, callback func() error, d time.Duration, attempts int) (err error) {
	for i := 0; i < attempts; i++ {
		err = callback()
		if err == nil {
			return nil
		}
		time.Sleep(d)
	}
	return err
}

// Logf writes logs to stdout if -v is set.
func Logf(str string, args ...interface{}) {
	if !testing.Verbose() {
		return
	}
	fmt.Printf(" %s | ", time.Now().Format("15:04:05"))
	fmt.Println(fmt.Sprintf(str, args...))
}
