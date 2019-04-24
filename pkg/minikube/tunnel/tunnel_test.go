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

	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

const RunningPid1 = 1234
const RunningPid2 = 1235
const NotRunningPid = 1236
const NotRunningPid2 = 1237

func mockPidChecker(pid int) (bool, error) {
	if pid == NotRunningPid || pid == NotRunningPid2 {
		return false, nil
	} else if pid == RunningPid1 || pid == RunningPid2 {
		return true, nil
	}
	return false, fmt.Errorf("fake pid checker does not recognize %d", pid)
}

type tunnelTestCase struct {
	name              string
	machineState      state.State
	serviceCIDR       string
	machineIP         string
	configLoaderError error
	mockPidHandling   bool
	call              func(tunnel *tunnel) (*Status, error)
	assertion         func(*testing.T, *Status, []*Status, []*Route, []*ID)
}

func simpleStopped() tunnelTestCase {
	return tunnelTestCase{
		name:         "simple stopped",
		machineState: state.Stopped,
		serviceCIDR:  "1.2.3.4/5",
		machineIP:    "1.2.3.4",
		call: func(tunnel *tunnel) (*Status, error) {
			return tunnel.update(), nil
		},
		assertion: func(t *testing.T, returnedState *Status, reportedStates []*Status, routes []*Route, registeredTunnels []*ID) {
			expectedState := &Status{
				MinikubeState: Stopped,
				MinikubeError: nil,
				TunnelID: ID{
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
			expectedReports := []*Status{returnedState}

			if !reflect.DeepEqual(reportedStates, expectedReports) {
				t.Errorf("wrong reports. expected %s\n got: %s", expectedReports, reportedStates)
			}
			if len(registeredTunnels) != 0 {
				t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
			}
		},
	}
}

func tunnelCleanupCtrlC() tunnelTestCase {
	return tunnelTestCase{
		name:         "tunnel cleanup (ctrl+c before check)",
		machineState: state.Running,
		serviceCIDR:  "1.2.3.4/5",
		machineIP:    "1.2.3.4",
		call: func(tunnel *tunnel) (*Status, error) {
			return tunnel.cleanup(), nil
		},
		assertion: func(t *testing.T, returnedState *Status, reportedStates []*Status, routes []*Route, registeredTunnels []*ID) {
			expectedState := &Status{
				MinikubeState: Running,
				MinikubeError: nil,
				TunnelID: ID{
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
	}
}

func tunnelCreateRoute() tunnelTestCase {
	return tunnelTestCase{
		name:         "tunnel create Route",
		machineState: state.Running,
		serviceCIDR:  "1.2.3.4/5",
		machineIP:    "1.2.3.4",
		call: func(tunnel *tunnel) (*Status, error) {
			return tunnel.update(), nil
		},
		assertion: func(t *testing.T, returnedState *Status, reportedStates []*Status, routes []*Route, registeredTunnels []*ID) {
			expectedRoute := unsafeParseRoute("1.2.3.4", "1.2.3.4/5")
			expectedState := &Status{
				MinikubeState: Running,
				MinikubeError: nil,
				TunnelID: ID{
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
			expectedReports := []*Status{returnedState}

			if !reflect.DeepEqual(reportedStates, expectedReports) {
				t.Errorf("wrong reports. expected %s\n got: %s", expectedReports, reportedStates)
			}

			if len(registeredTunnels) != 1 || !registeredTunnels[0].Equal(&expectedState.TunnelID) {
				t.Errorf("registry mismatch.\nexpected [%+v]\ngot     %+v", &expectedState.TunnelID, registeredTunnels)
			}
		},
	}
}

func tunnelCleanupErrorAfterSuccess() tunnelTestCase {
	return tunnelTestCase{
		name:         "tunnel cleanup error after 1 successful addRoute",
		machineState: state.Running,
		serviceCIDR:  "1.2.3.4/5",
		machineIP:    "1.2.3.4",
		call: func(tunnel *tunnel) (*Status, error) {
			tunnel.update()
			tunnel.router.(*fakeRouter).errorResponse = errors.New("testerror")
			return tunnel.cleanup(), nil
		},
		assertion: func(t *testing.T, actualSecondState *Status, reportedStates []*Status, routes []*Route, registeredTunnels []*ID) {
			expectedRoute := unsafeParseRoute("1.2.3.4", "1.2.3.4/5")
			expectedFirstState := &Status{
				MinikubeState: Running,
				MinikubeError: nil,
				TunnelID: ID{
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
				t.Errorf("wrong tunnel status. expected route error to contain '%s' \ngot:     %s", substring, actualSecondState.RouteError)
			}

			expectedRoutes := []*Route{expectedRoute}
			if !reflect.DeepEqual(routes, expectedRoutes) {
				t.Errorf("expected %s routes\n got: %s", expectedRoutes, routes)
			}

			expectedReports := []*Status{expectedFirstState}

			if !reflect.DeepEqual(reportedStates, expectedReports) {
				t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
			}
			if len(registeredTunnels) != 1 || !registeredTunnels[0].Equal(&expectedFirstState.TunnelID) {
				t.Errorf("registry mismatch.\nexpected [%+v]\ngot     %+v", &expectedFirstState.TunnelID, registeredTunnels)
			}
		},
	}
}

func tunnelCleanup() tunnelTestCase {
	return tunnelTestCase{
		name:         "tunnel cleanup",
		machineState: state.Running,
		serviceCIDR:  "1.2.3.4/5",
		machineIP:    "1.2.3.4",
		call: func(tunnel *tunnel) (*Status, error) {
			tunnel.update()
			return tunnel.cleanup(), nil
		},
		assertion: func(t *testing.T, actualSecondState *Status, reportedStates []*Status, routes []*Route, registeredTunnels []*ID) {
			expectedRoute := unsafeParseRoute("1.2.3.4", "1.2.3.4/5")
			expectedFirstState := &Status{
				MinikubeState: Running,
				MinikubeError: nil,
				TunnelID: ID{
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

			expectedReports := []*Status{expectedFirstState}

			if !reflect.DeepEqual(reportedStates, expectedReports) {
				t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
			}
			if len(registeredTunnels) > 0 {
				t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
			}
		},
	}
}

func raceCondition1() tunnelTestCase {
	return tunnelTestCase{
		name:            "race condition: other tunnel registers while in between routing and registration",
		machineState:    state.Running,
		serviceCIDR:     "1.2.3.4/5",
		machineIP:       "1.2.3.4",
		mockPidHandling: true,
		call: func(tunnel *tunnel) (*Status, error) {

			err := tunnel.registry.Register(&ID{
				Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
				MachineName: "testmachine",
				Pid:         RunningPid2,
			})
			if err != nil {
				return nil, err
			}
			tunnel.update()
			return tunnel.cleanup(), nil
		},
		assertion: func(t *testing.T, actualSecondState *Status, reportedStates []*Status, routes []*Route, registeredTunnels []*ID) {
			expectedRoute := unsafeParseRoute("1.2.3.4", "1.2.3.4/5")
			expectedFirstState := &Status{
				MinikubeState: Running,
				MinikubeError: nil,
				TunnelID: ID{
					Route:       expectedRoute,
					MachineName: "testmachine",
					Pid:         RunningPid1,
				},
				RouteError: errorTunnelAlreadyExists(&ID{
					Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
					MachineName: "testmachine",
					Pid:         RunningPid2,
				}),
			}

			if !reflect.DeepEqual(expectedFirstState, actualSecondState) {
				t.Errorf("wrong tunnel status.\nexpected %s\ngot:     %s", expectedFirstState, actualSecondState)
			}

			if len(routes) > 0 {
				t.Errorf("expected empty routes\n got: %s", routes)
			}

			expectedReports := []*Status{expectedFirstState}

			if !reflect.DeepEqual(reportedStates, expectedReports) {
				t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
			}
			if len(registeredTunnels) > 0 {
				t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
			}
		},
	}
}

func raceCondition2() tunnelTestCase {
	return tunnelTestCase{
		name:            "race condition: other tunnel registers and creates the same route first",
		machineState:    state.Running,
		serviceCIDR:     "1.2.3.4/5",
		machineIP:       "1.2.3.4",
		mockPidHandling: true,
		call: func(tunnel *tunnel) (*Status, error) {

			err := tunnel.registry.Register(&ID{
				Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
				MachineName: "testmachine",
				Pid:         RunningPid2,
			})
			if err != nil {
				return nil, err
			}
			tunnel.router.(*fakeRouter).rt = append(tunnel.router.(*fakeRouter).rt, routingTableLine{
				route: unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
				line:  "",
			})
			tunnel.update()
			return tunnel.cleanup(), nil
		},
		assertion: func(t *testing.T, actualSecondState *Status, reportedStates []*Status, routes []*Route, registeredTunnels []*ID) {
			expectedRoute := unsafeParseRoute("1.2.3.4", "1.2.3.4/5")
			expectedFirstState := &Status{
				MinikubeState: Running,
				MinikubeError: nil,
				TunnelID: ID{
					Route:       expectedRoute,
					MachineName: "testmachine",
					Pid:         RunningPid1,
				},
				RouteError: errorTunnelAlreadyExists(&ID{
					Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
					MachineName: "testmachine",
					Pid:         RunningPid2,
				}),
			}

			if !reflect.DeepEqual(expectedFirstState, actualSecondState) {
				t.Errorf("wrong tunnel status.\nexpected %s\ngot:     %s", expectedFirstState, actualSecondState)
			}

			if len(routes) > 0 {
				t.Errorf("expected empty routes\n got: %s", routes)
			}

			expectedReports := []*Status{expectedFirstState}

			if !reflect.DeepEqual(reportedStates, expectedReports) {
				t.Errorf("wrong reports.\nexpected %v\n\ngot:     %v", expectedReports, reportedStates)
			}
			if len(registeredTunnels) > 0 {
				t.Errorf("registry mismatch.\nexpected []\ngot     %+v", registeredTunnels)
			}
		},
	}
}

func TestTunnel(t *testing.T) {
	testCases := []tunnelTestCase{
		simpleStopped(),
		tunnelCleanupCtrlC(),
		tunnelCreateRoute(),
		tunnelCleanupErrorAfterSuccess(),
		tunnelCleanup(),
		raceCondition1(),
		raceCondition2(),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockPidHandling {
				origPidChecker := checkIfRunning
				checkIfRunning = mockPidChecker
				defer func() { checkIfRunning = origPidChecker }()

				origPidGetter := getPid
				getPid = func() int {
					return RunningPid1
				}
				defer func() { getPid = origPidGetter }()
			}
			machineName := "testmachine"
			machineAPI := &tests.MockAPI{
				FakeStore: tests.FakeStore{
					Hosts: map[string]*host.Host{
						machineName: {
							Driver: &tests.MockDriver{
								CurrentState: tc.machineState,
								IP:           tc.machineIP,
							},
						},
					},
				},
			}
			configLoader := &stubConfigLoader{
				c: &config.Config{
					KubernetesConfig: config.KubernetesConfig{
						ServiceCIDR: tc.serviceCIDR,
					}},
				e: tc.configLoaderError,
			}

			registry, cleanup := createTestRegistry(t)
			defer cleanup()

			tunnel, err := newTunnel(machineName, machineAPI, configLoader, newStubCoreClient(nil), registry, &fakeRouter{})
			if err != nil {
				t.Errorf("error creating tunnel: %s", err)
				return
			}
			reporter := &recordingReporter{}
			tunnel.reporter = reporter

			returnedState, err := tc.call(tunnel)
			if err != nil {
				t.Errorf("error registering tunnel: %s", err)
			}

			tunnels, err := registry.List()
			if err != nil {
				t.Errorf("error querying registry %s", err)
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
	store := &tests.MockAPI{
		FakeStore: tests.FakeStore{
			Hosts: map[string]*host.Host{
				machineName: {
					Driver: &tests.MockDriver{
						CurrentState: state.Stopped,
						IP:           "1.2.3.5",
					},
				},
			},
		},
	}

	configLoader := &stubConfigLoader{
		c: &config.Config{
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
		path: f.Name(),
	}

	_, err = newTunnel(machineName, store, configLoader, newStubCoreClient(nil), registry, &fakeRouter{})
	if err == nil || !strings.Contains(err.Error(), "error loading machine") {
		t.Errorf("expected error containing 'error loading machine', got %s", err)
	}
}
