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
	"time"

	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

// WaitForDefaultSA waits for the default service account to be created.
func WaitForDefaultSA(cs *kubernetes.Clientset, timeout time.Duration) error {
	klog.Info("waiting for default service account to be created ...")
	start := time.Now()
	saReady := func() (bool, error) {
		// equivalent to manual check of 'kubectl --context profile get serviceaccount default'
		sas, err := cs.CoreV1().ServiceAccounts("default").List(context.Background(), meta.ListOptions{})
		if err != nil {
			klog.Infof("temporary error waiting for default SA: %v", err)
			return false, nil
		}
		for _, sa := range sas.Items {
			if sa.Name == "default" {
				klog.Infof("found service account: %q", sa.Name)
				return true, nil
			}
		}
		return false, nil
	}
	if err := wait.PollImmediate(kconst.APICallRetryInterval, timeout, saReady); err != nil {
		return errors.Wrapf(err, "waited %s for SA", time.Since(start))
	}

	klog.Infof("duration metric: took %s for default service account to be created ...", time.Since(start))
	return nil
}
