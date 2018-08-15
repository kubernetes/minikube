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

///*
//Copyright 2018 The Kubernetes Authors All rights reserved.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//*/
//
package tunnel

import (
	"errors"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/sirupsen/logrus"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/minikube/tunnel/types"
	"reflect"
	"strings"
	"testing"
)

func TestTunnel(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	testCases := []struct {
		name              string
		machineState      state.State
		serviceCIDR       string
		machineIP         string
		configLoaderError error
		routerError       error
		expectedState     *types.TunnelState
		call              func(tunnel *minikubeTunnel) *types.TunnelState
		assertion         func(*testing.T, *types.TunnelState, []*types.TunnelState, []*types.Route)
	}{
		{
			name:         "simple stopped",
			machineState: state.Stopped,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *types.TunnelState {
				return tunnel.updateTunnelStatus()
			},
			assertion: func(t *testing.T, returnedState *types.TunnelState, reportedStates []*types.TunnelState, routes []*types.Route) {
				expectedState := &types.TunnelState{
					MinikubeState: types.Stopped,
					MinikubeError: nil,
					Route:         nil,
				}

				if !reflect.DeepEqual(expectedState, returnedState) {
					t.Errorf("wrong tunnel state. expected %s\n got: %s", expectedState, returnedState)
				}

				if len(routes) > 0 {
					t.Errorf("expected empty routes\n got: %s", routes)
				}
				expectedReports := []*types.TunnelState{returnedState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports. expected %s\n got: %s", expectedReports, reportedStates)
				}

			},
		},
		{
			name: "loading the machine config error",
			call: func(tunnel *minikubeTunnel) *types.TunnelState {
				return tunnel.updateTunnelStatus()
			},
			machineState:      state.None,
			configLoaderError: errors.New("error loading machine"),
			assertion: func(t *testing.T, returnedState *types.TunnelState, reportedStates []*types.TunnelState, routes []*types.Route) {
				if returnedState.MinikubeState != types.Unkown {
					t.Errorf("wrong tunnel state. expected Unkown\n got: %s", returnedState.MinikubeState)
				}

				if !strings.Contains(returnedState.MinikubeError.Error(), "error loading machine") {
					t.Errorf("wrong tunnel state. expected minikube error to contain 'error loading machine' \n got: %s", returnedState.MinikubeError)
				}

				if len(routes) > 0 {
					t.Errorf("expected empty routes\n got: %s", routes)
				}

				expectedReports := []*types.TunnelState{returnedState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports. expected %s\n got: %s", expectedReports, reportedStates)
				}
			},
		},
		{
			name:         "tunnel cleanup (ctrl+c before check)",
			machineState: state.Running,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *types.TunnelState {
				return tunnel.cleanup()
			},
			assertion: func(t *testing.T, returnedState *types.TunnelState, reportedStates []*types.TunnelState, routes []*types.Route) {
				expectedState := &types.TunnelState{
					MinikubeState: types.Unkown,
					MinikubeError: nil,
					Route:         nil,
				}

				if !reflect.DeepEqual(expectedState, returnedState) {
					t.Errorf("wrong tunnel state. expected %s\n got: %s", expectedState, returnedState)
				}

				if len(routes) > 0 {
					t.Errorf("expected empty routes\n got: %s", routes)
				}

				if len(reportedStates) > 0 {
					t.Errorf("wrong reports. expected no reports, got: %s", reportedStates)
				}
			},
		},

		{
			name:         "tunnel create Route",
			machineState: state.Running,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *types.TunnelState {
				return tunnel.updateTunnelStatus()
			},
			assertion: func(t *testing.T, returnedState *types.TunnelState, reportedStates []*types.TunnelState, routes []*types.Route) {
				expectedRoute := parseRoute("1.2.3.4", "1.2.3.4/5")
				expectedState := &types.TunnelState{
					MinikubeState: types.Running,
					MinikubeError: nil,
					Route:         expectedRoute,
				}

				if !reflect.DeepEqual(expectedState, returnedState) {
					t.Errorf("wrong tunnel state. expected %s\n got: %s", expectedState, returnedState)
				}

				expectedRoutes := []*types.Route{expectedRoute}
				if !reflect.DeepEqual(routes, expectedRoutes) {
					t.Errorf("expected %s routes\n got: %s", expectedRoutes, routes)
				}
				expectedReports := []*types.TunnelState{returnedState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports. expected %s\n got: %s", expectedReports, reportedStates)
				}
			},
		},
		{
			name:         "tunnel cleanup error after 1 successful update",
			machineState: state.Running,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *types.TunnelState {
				logrus.SetLevel(logrus.DebugLevel)
				defer logrus.SetLevel(logrus.InfoLevel)
				tunnel.updateTunnelStatus()
				tunnel.router.(*fakeRouter).errorResponse = errors.New("testerror")
				return tunnel.cleanup()
			},
			assertion: func(t *testing.T, actualSecondState *types.TunnelState, reportedStates []*types.TunnelState, routes []*types.Route) {
				expectedRoute := parseRoute("1.2.3.4", "1.2.3.4/5")
				expectedFirstState := &types.TunnelState{
					MinikubeState: types.Running,
					MinikubeError: nil,
					Route:         expectedRoute,
				}

				if actualSecondState.MinikubeState != types.Running {
					t.Errorf("wrong minikube state.\nexpected Running\ngot:     %s", actualSecondState.MinikubeState)
				}

				substring := "testerror"
				if !strings.Contains(actualSecondState.RouteError.Error(), substring) {
					t.Errorf("wrong tunnel state. expected Route error to contain '%s' \ngot:     %s", substring, actualSecondState.RouteError)
				}

				if !reflect.DeepEqual(actualSecondState.Route, expectedRoute) {
					t.Errorf("wrong Route in tunnel state. expected %s\n\ngot:     %s", expectedRoute, actualSecondState.Route)
				}

				expectedRoutes := []*types.Route{expectedRoute}
				if !reflect.DeepEqual(routes, expectedRoutes) {
					t.Errorf("expected %s routes\n got: %s", expectedRoutes, routes)
				}

				expectedReports := []*types.TunnelState{expectedFirstState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			machineName := "testmachine"
			store := &tests.FakeStore{
				Hosts: map[string]*host.Host{
					machineName: {
						Driver: &tests.MockDriver{
							CurrentState: tc.machineState,
							IP:           tc.machineIP,
						},
					},
				},
			}
			configLoader := &stubConfigLoader{
				c: config.Config{
					KubernetesConfig: config.KubernetesConfig{
						ServiceCIDR: tc.serviceCIDR,
					}},
				e: tc.configLoaderError,
			}
			tunnel := NewTunnel(machineName, store, configLoader, newStubCoreClient(nil, nil))
			tunnel.router = &fakeRouter{
				errorResponse: tc.routerError,
			}
			reporter := &recordingReporter{}
			tunnel.reporter = reporter

			returnedState := tc.call(tunnel)
			tc.assertion(t, returnedState, reporter.statesRecorded, tunnel.router.(*fakeRouter).osRoutes)

		})
	}

}
