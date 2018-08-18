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
	"github.com/docker/machine/libmachine/persist"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/minikube/pkg/minikube/config"
	"fmt"
	"os"
)

type tunnel interface {
	cleanup() *TunnelState
	updateTunnelStatus() *TunnelState
}

func newTunnel(machineName string,
	machineStore persist.Store,
	configLoader config.ConfigLoader,
	v1Core v1.CoreV1Interface, registry *persistentRegistry) (*minikubeTunnel, error) {
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
	err = registry.Register(&id)
	if err != nil {
		return nil, fmt.Errorf("error registering tunnel: %s", err)
	}

	return &minikubeTunnel{
		clusterInspector:     clusterInspector,
		router:               &osRouter{},
		registry:             registry,
		loadBalancerEmulator: NewLoadBalancerEmulator(v1Core),
		state: &TunnelState{
			TunnelID:      id,
			MinikubeState: state,
			MinikubeError: nil,
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

	state *TunnelState
}

func (t *minikubeTunnel) cleanup() *TunnelState {
	logrus.Debugf("cleaning up %s", t.state.TunnelID.Route)
	e := t.router.Cleanup(t.state.TunnelID.Route)
	if e != nil {
		t.state.RouterError = errors.Errorf("error cleaning up Route %s", e)
	}
	if t.state.MinikubeState == Running {
		t.state.PatchedServices, t.state.LoadBalancerEmulatorError = t.loadBalancerEmulator.Cleanup()
	}
	t.registry.Remove(t.state.TunnelID.Route)
	return t.state
}

func (t *minikubeTunnel) updateTunnelStatus() *TunnelState {
	logrus.Debug("updating tunnel status...")
	t.state.MinikubeState, _, t.state.MinikubeError = t.clusterInspector.getStateAndHost()
	if t.state.MinikubeState == Running {
		logrus.Debug("minikube is running, trying to add Route %s", t.state.TunnelID.Route)

		t.state.RouterError = t.router.EnsureRouteIsAdded(t.state.TunnelID.Route)
		t.state.PatchedServices, t.state.LoadBalancerEmulatorError = t.loadBalancerEmulator.PatchServices()
	}
	logrus.Debugf("sending report %v", t.state)
	t.reporter.Report(t.state.Clone())
	return t.state
}
