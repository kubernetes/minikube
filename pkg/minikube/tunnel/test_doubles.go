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
	"github.com/sirupsen/logrus"
	"k8s.io/minikube/pkg/minikube/config"
)

type recordingReporter struct {
	statesRecorded [] *TunnelState
}

func (r *recordingReporter) Report(tunnelState  *TunnelState) {
	logrus.Debugf("recordingReporter.Report: %v", tunnelState)
	r.statesRecorded = append(r.statesRecorded, tunnelState)
}

type fakeRouter struct {
	osRoutes      [] *Route
	osRouter
	errorResponse error
}

func (r *fakeRouter) EnsureRouteIsAdded(route  *Route) error {
	if r.errorResponse == nil {
		r.osRoutes = append(r.osRoutes, route)
	}
	return r.errorResponse
}
func (r *fakeRouter) Cleanup(route  *Route) error {
	logrus.Infof("fake router cleanup: %v\n", route)
	if r.errorResponse == nil {
		for i := range r.osRoutes {
			if r.osRoutes[i].Equal(route) {
				r.osRoutes = append(r.osRoutes[:i], r.osRoutes[i+1:]...)
				break
			}
		}
	}
	return r.errorResponse
}

type stubConfigLoader struct {
	c config.Config
	e error
}

func (l *stubConfigLoader) LoadConfigFromFile(profile string) (config.Config, error) {
	return l.c, l.e
}

type fakeTunnelRegistry struct {
	tunnels map[ *Route]*TunnelID
}

func (r *fakeTunnelRegistry) Register(route *TunnelID) error            {
	r.tunnels[route.Route] = route
	return nil
}
func (r *fakeTunnelRegistry) Get(route  *Route) (*TunnelID, error) {
	return r.tunnels[route], nil
}
func (r *fakeTunnelRegistry) Remove(route  *Route) error                    {
	return nil
}
func (r *fakeTunnelRegistry) List() []*TunnelID {
	return nil
}
