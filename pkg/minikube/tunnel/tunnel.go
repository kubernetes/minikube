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
	"k8s.io/minikube/pkg/minikube/tunnel/types"
)

type Tunnel interface {
	cleanup() *types.TunnelState
	updateTunnelStatus() *types.TunnelState
}

func NewTunnel(machineName string,
	machineStore persist.Store,
	configLoader config.ConfigLoader,
	v1Core v1.CoreV1Interface) *minikubeTunnel {
	return &minikubeTunnel{
		clusterInspector: &minikubeInspector{
			machineName:  machineName,
			machineStore: machineStore,
			configLoader: configLoader,
		},
		router:              &osRouter{},
		loadBalancerPatcher: NewLoadBalancerPatcher(v1Core),
		state: &types.TunnelState{
			MinikubeState: types.Unkown,
			MinikubeError: nil,
		},
	}
}

type minikubeTunnel struct {
	//collaborators
	clusterInspector    *minikubeInspector
	router              router
	loadBalancerPatcher *loadBalancerPatcher
	state               *types.TunnelState
	reporter            reporter
}

func (t *minikubeTunnel) cleanup() *types.TunnelState {
	if t.state != nil && t.state.Route != nil {
		logrus.Debugf("cleaning up %s", t.state.Route)
		e := t.router.Cleanup(t.state.Route)
		if e != nil {
			t.state.RouteError = errors.Errorf("error cleaning up Route %s", e)
		} else {
			t.state.Route = nil
		}
	}
	if t.state.MinikubeState == types.Running {
		t.state.PatchedServices, t.state.LoadBalancerPatcherError = t.loadBalancerPatcher.Cleanup()
	}
	return t.state
}

func (t *minikubeTunnel) updateTunnelStatus() *types.TunnelState {
	logrus.Debug("updating tunnel status...")
	s, r, e := t.clusterInspector.Inspect()
	//TODO: can this change while minikube is running?
	//TODO: check for registered tunnels
	t.state = &types.TunnelState{
		MinikubeState: s,
		MinikubeError: e,
		Route:         r,
	}
	if t.state.MinikubeState == types.Running {
		logrus.Debug("minikube is running trying to add Route %s", t.state.Route)
		t.state.RouteError = t.router.EnsureRouteIsAdded(t.state.Route)
		t.state.PatchedServices, t.state.LoadBalancerPatcherError = t.loadBalancerPatcher.PatchServices()
	}
	logrus.Debugf("sending report %v", t.state)
	t.reporter.Report(t.state.Clone())
	return t.state
}
