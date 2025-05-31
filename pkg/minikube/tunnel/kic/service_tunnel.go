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

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/klog/v2"
)

// ServiceTunnel manages an SSH tunnel for a Kubernetes service.
// It holds configuration for the SSH connection and the tunnel's state.
type ServiceTunnel struct {
	sshPort        string
	sshKey         string
	v1Core         typed_core.CoreV1Interface
	sshConn        *sshConn
	suppressStdOut bool
}

// NewServiceTunnel creates and returns a new ServiceTunnel instance.
// sshPort is the port number for the SSH connection.
// sshKey is the path to the SSH private key file.
// v1Core is the Kubernetes CoreV1 client interface for interacting with services.
// suppressStdOut controls whether standard output from the tunnel process should be suppressed.
func NewServiceTunnel(sshPort, sshKey string, v1Core typed_core.CoreV1Interface, suppressStdOut bool) *ServiceTunnel {
	return &ServiceTunnel{
		sshPort:        sshPort,
		sshKey:         sshKey,
		v1Core:         v1Core,
		suppressStdOut: suppressStdOut,
	}
}

// Start establishes an SSH tunnel for the specified Kubernetes service.
// It retrieves service details, creates an SSH connection with random local ports
// for each service port, and starts the tunnel in a new goroutine.
// It returns a slice of URLs (e.g., "http://127.0.0.1:local_port") corresponding
// to the tunnelled ports, or an error if the setup fails.
// Errors from the tunnel running in the background are logged via klog.
func (t *ServiceTunnel) Start(svcName, namespace string) ([]string, error) {
	svc, err := t.v1Core.Services(namespace).Get(context.Background(), svcName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Service %s was not found in %q namespace. You may select another namespace by using 'minikube service %s -n <namespace>", svcName, namespace, svcName)
	}

	t.sshConn, err = createSSHConnWithRandomPorts(svcName, t.sshPort, t.sshKey, svc)
	if err != nil {
		return nil, errors.Wrap(err, "creating ssh conn")
	}

	go func() {
		t.sshConn.suppressStdOut = t.suppressStdOut
		err = t.sshConn.startAndWait()
		if err != nil {
			klog.Errorf("error starting ssh tunnel: %v", err)
		}
	}()

	urls := make([]string, 0, len(svc.Spec.Ports))
	for _, port := range t.sshConn.ports {
		urls = append(urls, fmt.Sprintf("http://127.0.0.1:%d", port))
	}

	return urls, nil
}

// Stop attempts to gracefully stop the active SSH tunnel.
// Any errors encountered during the stop process are logged as warnings.
func (t *ServiceTunnel) Stop() {
	err := t.sshConn.stop()
	if err != nil {
		klog.Warningf("Failed to stop ssh tunnel: %v", err)
	}
}
