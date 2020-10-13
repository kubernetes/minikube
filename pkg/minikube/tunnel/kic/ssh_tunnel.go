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

package kic

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/tunnel"
)

// SSHTunnel ...
type SSHTunnel struct {
	ctx                  context.Context
	sshPort              string
	sshKey               string
	v1Core               typed_core.CoreV1Interface
	LoadBalancerEmulator tunnel.LoadBalancerEmulator
	conns                map[string]*sshConn
	connsToStop          map[string]*sshConn
}

// NewSSHTunnel ...
func NewSSHTunnel(ctx context.Context, sshPort, sshKey string, v1Core typed_core.CoreV1Interface) *SSHTunnel {
	return &SSHTunnel{
		ctx:                  ctx,
		sshPort:              sshPort,
		sshKey:               sshKey,
		v1Core:               v1Core,
		LoadBalancerEmulator: tunnel.NewLoadBalancerEmulator(v1Core),
		conns:                make(map[string]*sshConn),
		connsToStop:          make(map[string]*sshConn),
	}
}

// Start ...
func (t *SSHTunnel) Start() error {
	for {
		select {
		case <-t.ctx.Done():
			_, err := t.LoadBalancerEmulator.Cleanup()
			if err != nil {
				klog.Errorf("error cleaning up: %v", err)
			}
			return err
		default:
		}

		services, err := t.v1Core.Services("").List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("error listing services: %v", err)
		}

		t.markConnectionsToBeStopped()

		for _, svc := range services.Items {
			if svc.Spec.Type == v1.ServiceTypeLoadBalancer {
				t.startConnection(svc)
			}
		}

		t.stopMarkedConnections()

		// TODO: which time to use?
		time.Sleep(1 * time.Second)
	}
}

func (t *SSHTunnel) markConnectionsToBeStopped() {
	for _, conn := range t.conns {
		t.connsToStop[conn.name] = conn
	}
}

func (t *SSHTunnel) startConnection(svc v1.Service) {
	uniqName := sshConnUniqName(svc)
	existingSSHConn, ok := t.conns[uniqName]

	if ok {
		// if the svc still exist we remove the conn from the stopping list
		delete(t.connsToStop, existingSSHConn.name)
		return
	}

	// create new ssh conn
	newSSHConn := createSSHConn(uniqName, t.sshPort, t.sshKey, &svc)
	t.conns[newSSHConn.name] = newSSHConn

	go func() {
		err := newSSHConn.startAndWait()
		if err != nil {
			klog.Errorf("error starting ssh tunnel: %v", err)
		}
	}()

	err := t.LoadBalancerEmulator.PatchServiceIP(t.v1Core.RESTClient(), svc, "127.0.0.1")
	if err != nil {
		klog.Errorf("error patching service: %v", err)
	}
}

func (t *SSHTunnel) stopMarkedConnections() {
	for _, sshConn := range t.connsToStop {
		err := sshConn.stop()
		if err != nil {
			klog.Errorf("error stopping ssh tunnel: %v", err)
		}
		delete(t.conns, sshConn.name)
		delete(t.connsToStop, sshConn.name)
	}
}

// sshConnName creates a uniq name for the tunnel, using its name/clusterIP/ports.
// This allows a new process to be created if an existing service was changed,
// the new process will support the IP/Ports change occurred.
func sshConnUniqName(service v1.Service) string {
	n := []string{
		service.Name,
		"-",
		service.Spec.ClusterIP,
	}

	for _, port := range service.Spec.Ports {
		n = append(n, fmt.Sprintf("-%d", port.Port))
	}

	return strings.Join(n, "")
}
