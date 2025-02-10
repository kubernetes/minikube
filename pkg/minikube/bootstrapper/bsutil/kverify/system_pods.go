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
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/util/retry"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

// WaitForSystemPods verifies essential pods for running kurnetes is running
func WaitForSystemPods(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner, client *kubernetes.Clientset, start time.Time, timeout time.Duration) error {
	klog.Info("waiting for kube-system pods to appear ...")
	pStart := time.Now()

	podList := func() error {
		if time.Since(start) > minLogCheckTime {
			announceProblems(r, bs, cfg, cr)
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		// Wait for any system pod, as waiting for apiserver may block until etcd
		pods, err := client.CoreV1().Pods("kube-system").List(context.Background(), meta.ListOptions{})
		if err != nil {
			klog.Warningf("pod list returned error: %v", err)
			return err
		}

		klog.Infof("%d kube-system pods found", len(pods.Items))
		for _, pod := range pods.Items {
			klog.Info(podStatusMsg(pod))
		}

		if len(pods.Items) < 2 {
			return fmt.Errorf("only %d pod(s) have shown up", len(pods.Items))
		}

		return nil
	}

	if err := retry.Local(podList, timeout); err != nil {
		return fmt.Errorf("apiserver never returned a pod list")
	}
	klog.Infof("duration metric: took %s to wait for pod list to return data ...", time.Since(pStart))
	return nil
}

// ExpectAppsRunning returns whether or not all expected k8s-apps are running. (without waiting for them)
func ExpectAppsRunning(cs *kubernetes.Clientset, expected []string) error {
	found := map[string]bool{}

	pods, err := cs.CoreV1().Pods("kube-system").List(context.Background(), meta.ListOptions{})
	if err != nil {
		return err
	}
	klog.Infof("%d kube-system pods found", len(pods.Items))

	for _, pod := range pods.Items {
		klog.Info(podStatusMsg(pod))

		if pod.Status.Phase != core.PodRunning {
			continue
		}

		for k, v := range pod.ObjectMeta.Labels {
			if k == "component" || k == "k8s-app" {
				found[v] = true
			}
		}
	}

	missing := []string{}
	for _, e := range expected {
		if !found[e] {
			missing = append(missing, e)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing components: %v", strings.Join(missing, ", "))
	}
	return nil
}

// WaitForAppsRunning waits for expected Apps To be running
func WaitForAppsRunning(cs *kubernetes.Clientset, expected []string, timeout time.Duration) error {
	klog.Info("waiting for k8s-apps to be running ...")
	start := time.Now()

	checkRunning := func() error {
		return ExpectAppsRunning(cs, expected)
	}

	if err := retry.Local(checkRunning, timeout); err != nil {
		return errors.Wrapf(err, "expected k8s-apps")
	}
	klog.Infof("duration metric: took %s to wait for k8s-apps to be running ...", time.Since(start))
	return nil
}

// podStatusMsg returns a human-readable pod status, for generating debug status
func podStatusMsg(pod core.Pod) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%q [%s] %s", pod.ObjectMeta.GetName(), pod.ObjectMeta.GetUID(), pod.Status.Phase))
	for i, c := range pod.Status.Conditions {
		if c.Reason != "" {
			if i == 0 {
				sb.WriteString(": ")
			} else {
				sb.WriteString(" / ")
			}
			sb.WriteString(fmt.Sprintf("%s:%s", c.Type, c.Reason))
		}
		if c.Message != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", c.Message))
		}
	}
	return sb.String()
}

// announceProblems checks for problems, and slows polling down if any are found
func announceProblems(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner) {
	problems := logs.FindProblems(r, bs, cfg, cr)
	if len(problems) > 0 {
		logs.OutputProblems(problems, 5, os.Stderr)
		time.Sleep(kconst.APICallRetryInterval * 15)
	}
}
