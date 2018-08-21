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
	"errors"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/sirupsen/logrus"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/tests"

	"io/ioutil"
	"os"
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
		expectedState     *TunnelStatus
		call              func(tunnel *minikubeTunnel) *TunnelStatus
		assertion         func(*testing.T, *TunnelStatus, []*TunnelStatus, []*Route, []*TunnelID)
	}{
		{
			name:         "simple stopped",
			machineState: state.Stopped,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *TunnelStatus {
				return tunnel.updateTunnelStatus()
			},
			assertion: func(t *testing.T, returnedState *TunnelStatus, reportedStates []*TunnelStatus, routes []*Route, registeredTunnels []*TunnelID) {
				expectedState := &TunnelStatus{
					MinikubeState: Stopped,
					MinikubeError: nil,
					TunnelID: TunnelID{
						Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
						MachineName: "testmachine",
						Pid:         os.Getpid(),
					},
				}

				if !reflect.DeepEqual(expectedState, returnedState) {
					t.Errorf("wrong tunnel status.\nexpected %s\ngot:     %s", expectedState, returnedState)
				}

				if len(routes) > 0 {
					t.Errorf("expected empty routes\n got: %s", routes)
				}
				expectedReports := []*TunnelStatus{returnedState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports. expected %s\n got: %s", expectedReports, reportedStates)
				}
				if len(registeredTunnels) != 0 {
					t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
				}
			},
		},
		{
			name:         "tunnel cleanup (ctrl+c before check)",
			machineState: state.Running,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *TunnelStatus {
				return tunnel.cleanup()
			},
			assertion: func(t *testing.T, returnedState *TunnelStatus, reportedStates []*TunnelStatus, routes []*Route, registeredTunnels []*TunnelID) {
				expectedState := &TunnelStatus{
					MinikubeState: Running,
					MinikubeError: nil,
					TunnelID: TunnelID{
						Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
						MachineName: "testmachine",
						Pid:         os.Getpid(),
					},
				}

				if !reflect.DeepEqual(expectedState, returnedState) {
					t.Errorf("wrong tunnel status.\nexpected %s\ngot:     %s", expectedState, returnedState)
				}

				if len(routes) > 0 {
					t.Errorf("expected empty routes\n got: %s", routes)
				}

				if len(reportedStates) > 0 {
					t.Errorf("wrong reports. expected no reports, got: %s", reportedStates)
				}

				if len(registeredTunnels) > 0 {
					t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
				}
			},
		},

		{
			name:         "tunnel create Route",
			machineState: state.Running,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *TunnelStatus {
				return tunnel.updateTunnelStatus()
			},
			assertion: func(t *testing.T, returnedState *TunnelStatus, reportedStates []*TunnelStatus, routes []*Route, registeredTunnels []*TunnelID) {
				expectedRoute := unsafeParseRoute("1.2.3.4", "1.2.3.4/5")
				expectedState := &TunnelStatus{
					MinikubeState: Running,
					MinikubeError: nil,
					TunnelID: TunnelID{
						Route:       expectedRoute,
						MachineName: "testmachine",
						Pid:         os.Getpid(),
					},
				}

				if !reflect.DeepEqual(expectedState, returnedState) {
					t.Errorf("wrong tunnel status. expected %s\n got: %s", expectedState, returnedState)
				}

				expectedRoutes := []*Route{expectedRoute}
				if !reflect.DeepEqual(routes, expectedRoutes) {
					t.Errorf("expected %s routes\n got: %s", expectedRoutes, routes)
				}
				expectedReports := []*TunnelStatus{returnedState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports. expected %s\n got: %s", expectedReports, reportedStates)
				}

				if len(registeredTunnels) != 1 || !registeredTunnels[0].Equal(&expectedState.TunnelID) {
					t.Errorf("registry mismatch.\nexpected [%+v]\ngot     %+v", &expectedState.TunnelID, registeredTunnels)
				}
			},
		},
		{
			name:         "tunnel cleanup error after 1 successful update",
			machineState: state.Running,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *TunnelStatus {
				logrus.SetLevel(logrus.DebugLevel)
				defer logrus.SetLevel(logrus.InfoLevel)
				tunnel.updateTunnelStatus()
				tunnel.router.(*fakeRouter).errorResponse = errors.New("testerror")
				return tunnel.cleanup()
			},
			assertion: func(t *testing.T, actualSecondState *TunnelStatus, reportedStates []*TunnelStatus, routes []*Route, registeredTunnels []*TunnelID) {
				expectedRoute := unsafeParseRoute("1.2.3.4", "1.2.3.4/5")
				expectedFirstState := &TunnelStatus{
					MinikubeState: Running,
					MinikubeError: nil,
					TunnelID: TunnelID{
						Route:       expectedRoute,
						MachineName: "testmachine",
						Pid:         os.Getpid(),
					},
				}

				if actualSecondState.MinikubeState != Running {
					t.Errorf("wrong minikube status.\nexpected Running\ngot:     %s", actualSecondState.MinikubeState)
				}

				substring := "testerror"
				if !strings.Contains(actualSecondState.RouterError.Error(), substring) {
					t.Errorf("wrong tunnel status. expected Route error to contain '%s' \ngot:     %s", substring, actualSecondState.RouterError)
				}

				expectedRoutes := []*Route{expectedRoute}
				if !reflect.DeepEqual(routes, expectedRoutes) {
					t.Errorf("expected %s routes\n got: %s", expectedRoutes, routes)
				}

				expectedReports := []*TunnelStatus{expectedFirstState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
				}
				if len(registeredTunnels) != 1 || !registeredTunnels[0].Equal(&expectedFirstState.TunnelID) {
					t.Errorf("registry mismatch.\nexpected [%+v]\ngot     %+v", &expectedFirstState.TunnelID, registeredTunnels)
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

			registry, cleanup := createTestRegistry(t)
			defer cleanup()

			tunnel, e := newTunnel(machineName, store, configLoader, newStubCoreClient(nil, nil), registry, &fakeRouter{
				errorResponse: tc.routerError,
			})
			if e != nil {
				t.Errorf("error creating tunnel: %s", e)
				return
			}
			reporter := &recordingReporter{}
			tunnel.reporter = reporter

			returnedState := tc.call(tunnel)
			tunnels, e := registry.List()
			if e != nil {
				t.Errorf("error querying registry %s", e)
				return
			}
			var routes []*Route
			for _, r := range tunnel.router.(*fakeRouter).rt {
				routes = append(routes, r.route)
			}
			tc.assertion(t, returnedState, reporter.statesRecorded, routes, tunnels)

		})
	}

}

func TestErrorCreatingTunnel(t *testing.T) {
	machineName := "testmachine"
	store := &tests.FakeStore{
		Hosts: map[string]*host.Host{
			machineName: {
				Driver: &tests.MockDriver{
					CurrentState: state.Stopped,
					IP:           "1.2.3.5",
				},
			},
		},
	}
	configLoader := &stubConfigLoader{
		c: config.Config{
			KubernetesConfig: config.KubernetesConfig{
				ServiceCIDR: "10.96.0.0/12",
			}},
		e: errors.New("error loading machine"),
	}

	f, err := ioutil.TempFile(os.TempDir(), "reg_")
	f.Close()
	if err != nil {
		t.Errorf("failed to create temp file %s", err)
	}
	defer os.Remove(f.Name())
	registry := &persistentRegistry{
		fileName: f.Name(),
	}

	_, e := newTunnel(machineName, store, configLoader, newStubCoreClient(nil, nil), registry, &fakeRouter{})
	if e == nil || !strings.Contains(e.Error(), "error loading machine") {
		t.Errorf("expected error containing 'error loading machine', got %s", e)
	}
}
