/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"fmt"
	"github.com/docker/machine/libmachine/persist"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/minikube/pkg/minikube/config"
	"os"
)

type tunnel interface {
	cleanup() *TunnelStatus
	updateTunnelStatus() *TunnelStatus
}

func newTunnel(machineName string,
	machineStore persist.Store,
	configLoader config.ConfigLoader,
	v1Core v1.CoreV1Interface, registry *persistentRegistry, router router) (*minikubeTunnel, error) {
	clusterInspector := &minikubeInspector{
		machineName:  machineName,
		machineStore: machineStore,
		configLoader: configLoader,
	}
	state, route, err := clusterInspector.getStateAndRoute()
	if err != nil {
		return nil, fmt.Errorf("unable to determine cluster info: %s", err)
	}
	id := TunnelID{
		Route:       route,
		MachineName: machineName,
		Pid:         os.Getpid(),
	}

	return &minikubeTunnel{
		clusterInspector:     clusterInspector,
		router:               router,
		registry:             registry,
		loadBalancerEmulator: NewLoadBalancerEmulator(v1Core),
		status: &TunnelStatus{
			TunnelID:      id,
			MinikubeState: state,
		},
		reporter: &simpleReporter{
			out: os.Stdout,
		},
	}, nil

}

type minikubeTunnel struct {
	//collaborators
	clusterInspector     *minikubeInspector
	router               router
	loadBalancerEmulator *loadBalancerEmulator
	reporter             reporter
	registry             *persistentRegistry

	status *TunnelStatus
}

func (t *minikubeTunnel) cleanup() *TunnelStatus {
	logrus.Debugf("cleaning up %s", t.status.TunnelID.Route)
	e := t.router.Cleanup(t.status.TunnelID.Route)
	if e != nil {
		t.status.RouterError = errors.Errorf("error cleaning up route: %s", e)
		logrus.Debugf(t.status.RouterError.Error())
	} else {
		exists, _, _, err := t.router.Inspect(t.status.TunnelID.Route)
		if !exists && err != nil {
			t.registry.Remove(t.status.TunnelID.Route)
		} else {
			logrus.Debugf("did not remove tunnel (%s) from registry.")
		}
	}
	if t.status.MinikubeState == Running {
		t.status.PatchedServices, t.status.LoadBalancerEmulatorError = t.loadBalancerEmulator.Cleanup()
	}
	return t.status
}

func (t *minikubeTunnel) updateTunnelStatus() *TunnelStatus {
	logrus.Debug("updating tunnel status...")
	t.status.MinikubeState, _, t.status.MinikubeError = t.clusterInspector.getStateAndHost()
	if t.status.MinikubeState == Running {
		logrus.Debug("minikube is running, trying to add Route %s", t.status.TunnelID.Route)

		exists, conflict, _, err := t.router.Inspect(t.status.TunnelID.Route)
		if err != nil {
			t.status.RouterError = fmt.Errorf("error checking for route state: %s", err)
		} else if !exists && len(conflict) == 0 {
			t.status.RouterError = t.router.EnsureRouteIsAdded(t.status.TunnelID.Route)
			if t.status.RouterError == nil {
				//the route was added successfully, we need to make sure the registry has it too
				if err := t.registry.Register(&t.status.TunnelID); err != nil {
					logrus.Errorf("failed to register tunnel: %s, removing...", err)
					//if registry failes, we need to remove the route
					t.status.RouterError = t.router.Cleanup(t.status.TunnelID.Route)
				}
			}
		} else if len(conflict) > 0 {
			t.status.RouterError = fmt.Errorf("conflicting route: ")
		}
		t.status.PatchedServices, t.status.LoadBalancerEmulatorError = t.loadBalancerEmulator.PatchServices()
	}
	logrus.Debugf("sending report %s", t.status)
	t.reporter.Report(t.status.Clone())
	return t.status
}
