/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package kubeadm

import (
	"context"
	"fmt"
	"time"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
)

// waitForWorkerRegistration waits for the exact joined worker before labels are applied.
// It uses the existing control-plane client and a finite registration budget.
func (k *Bootstrapper) waitForWorkerRegistration(cfg config.ClusterConfig, name string) error {
	cp, err := config.ControlPlane(cfg)
	if err != nil {
		return fmt.Errorf("get control-plane node: %w", err)
	}
	hostname, _, port, err := driver.ControlPlaneEndpoint(&cfg, &cp, cfg.Driver)
	if err != nil {
		return fmt.Errorf("get control-plane endpoint: %w", err)
	}
	c, err := k.client(hostname, port)
	if err != nil {
		return fmt.Errorf("kubernetes client: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	klog.Infof("Waiting for joined worker registration before labeling")
	return waitForRegisteredNode(ctx, c.CoreV1().Nodes(), name, time.Second)
}

// waitForRegisteredNode observes presence, independently of the later NodeReady check.
// Missing nodes and transient API pressure retry within ctx; permission and other
// permanent failures return immediately. The function performs read-only requests.
func waitForRegisteredNode(ctx context.Context, nodes coreclient.NodeInterface, name string, interval time.Duration) error {
	return wait.PollUntilContextCancel(ctx, interval, true, func(ctx context.Context) (bool, error) {
		_, err := nodes.Get(ctx, name, meta.GetOptions{})
		switch {
		case apierr.IsNotFound(err),
			apierr.IsTimeout(err),
			apierr.IsServerTimeout(err),
			apierr.IsServiceUnavailable(err),
			apierr.IsTooManyRequests(err):
			return false, nil
		}
		return err == nil, err
	})
}
