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
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/tests"

	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"fmt"
)

func TestTunnel(t *testing.T) {
	const RUNNING_PID1 = 1234
	const RUNNING_PID2 = 1235
	const NOT_RUNNING_PID = 1236
	const NOT_RUNNING_PID2 = 1237

	mockPidChecker := func(pid int) (bool, error) {
		if pid == NOT_RUNNING_PID || pid == NOT_RUNNING_PID2 {
			return false, nil
		} else if pid == RUNNING_PID1 || pid == RUNNING_PID2 {
			return true, nil
		}
		return false, fmt.Errorf("fake pid checker does not recognize %d", pid)
	}

	testCases := []struct {
		name              string
		machineState      state.State
		serviceCIDR       string
		machineIP         string
		configLoaderError error
		mockPidHandling   bool
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
				if actualSecondState.RouteError == nil || !strings.Contains(actualSecondState.RouteError.Error(), substring) {
					t.Errorf("wrong tunnel status. expected Route error to contain '%s' \ngot:     %s", substring, actualSecondState.RouteError)
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
		{
			name:         "tunnel cleanup",
			machineState: state.Running,
			serviceCIDR:  "1.2.3.4/5",
			machineIP:    "1.2.3.4",
			call: func(tunnel *minikubeTunnel) *TunnelStatus {
				tunnel.updateTunnelStatus()
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

				if !reflect.DeepEqual(expectedFirstState, actualSecondState) {
					t.Errorf("wrong tunnel status.\nexpected %s\ngot:     %s", expectedFirstState, actualSecondState)
				}

				if len(routes) > 0 {
					t.Errorf("expected empty routes\n got: %s", routes)
				}

				expectedReports := []*TunnelStatus{expectedFirstState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
				}
				if len(registeredTunnels) > 0 {
					t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
				}
			},
		},
		{
			name:            "race condition: other tunnel registers while in between routing and registration",
			machineState:    state.Running,
			serviceCIDR:     "1.2.3.4/5",
			machineIP:       "1.2.3.4",
			mockPidHandling: true,
			call: func(tunnel *minikubeTunnel) *TunnelStatus {

				tunnel.registry.Register(&TunnelID{
					Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
					MachineName: "testmachine",
					Pid:         RUNNING_PID2,
				}, )
				tunnel.updateTunnelStatus()
				return tunnel.cleanup()
			},
			assertion: func(t *testing.T, actualSecondState *TunnelStatus, reportedStates []*TunnelStatus, routes []*Route, registeredTunnels []*TunnelID) {
				b, e := checkIfRunning(1)
				fmt.Println(fmt.Sprintf("PID1: %v, %s", b, e))
				expectedRoute := unsafeParseRoute("1.2.3.4", "1.2.3.4/5")
				expectedFirstState := &TunnelStatus{
					MinikubeState: Running,
					MinikubeError: nil,
					TunnelID: TunnelID{
						Route:       expectedRoute,
						MachineName: "testmachine",
						Pid:         RUNNING_PID1,
					},
					RouteError: errorTunnelAlreadyExists(&TunnelID{
						Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
						MachineName: "testmachine",
						Pid:         RUNNING_PID2,
					}),
				}

				if !reflect.DeepEqual(expectedFirstState, actualSecondState) {
					t.Errorf("wrong tunnel status.\nexpected %s\ngot:     %s", expectedFirstState, actualSecondState)
				}

				if len(routes) > 0 {
					t.Errorf("expected empty routes\n got: %s", routes)
				}

				expectedReports := []*TunnelStatus{expectedFirstState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
				}
				if len(registeredTunnels) > 0 {
					t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
				}
			},
		},
		{
			name:            "race condition: other tunnel registers and creates the same route first",
			machineState:    state.Running,
			serviceCIDR:     "1.2.3.4/5",
			machineIP:       "1.2.3.4",
			mockPidHandling: true,
			call: func(tunnel *minikubeTunnel) *TunnelStatus {

				tunnel.registry.Register(&TunnelID{
					Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
					MachineName: "testmachine",
					Pid:         RUNNING_PID2,
				}, )
				tunnel.router.(*fakeRouter).rt = append(tunnel.router.(*fakeRouter).rt, routingTableLine{
					route: unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
					line:  "",
				})
				tunnel.updateTunnelStatus()
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
						Pid:         RUNNING_PID1,
					},
					RouteError: errorTunnelAlreadyExists(&TunnelID{
						Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
						MachineName: "testmachine",
						Pid:         RUNNING_PID2,
					}),
				}

				if !reflect.DeepEqual(expectedFirstState, actualSecondState) {
					t.Errorf("wrong tunnel status.\nexpected %s\ngot:     %s", expectedFirstState, actualSecondState)
				}

				if len(routes) > 0 {
					t.Errorf("expected empty routes\n got: %s", routes)
				}

				expectedReports := []*TunnelStatus{expectedFirstState}

				if !reflect.DeepEqual(reportedStates, expectedReports) {
					t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
				}
				if len(registeredTunnels) > 0 {
					t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
				}
			},
		},

	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockPidHandling {
				origPidChecker := checkIfRunning
				checkIfRunning = mockPidChecker
				defer func() { checkIfRunning = origPidChecker }()

				origPidGetter := getPid
				getPid = func() int {
					return RUNNING_PID1
				}
				defer func() { getPid = origPidGetter }()
			}
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

			tunnel, e := newTunnel(machineName, store, configLoader, newStubCoreClient(nil, nil), registry, &fakeRouter{})
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
