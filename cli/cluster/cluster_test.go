/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"fmt"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/cli/constants"
	"k8s.io/minikube/cli/tests"
)

func TestCreateHost(t *testing.T) {
	api := &tests.MockAPI{}

	exists, _ := api.Exists(constants.MachineName)
	if exists {
		t.Fatal("Machine already exists.")
	}
	_, err := createHost(api)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	exists, _ = api.Exists(constants.MachineName)
	if !exists {
		t.Fatal("Machine does not exist, but should.")
	}

	h, err := api.Load(constants.MachineName)
	if err != nil {
		t.Fatalf("Error loading machine: %v", err)
	}

	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine is not running. State is: %s", s)
	}

	if h.DriverName != "virtualbox" {
		t.Fatalf("Wrong driver name: %v. Should be virtualbox.", h.DriverName)
	}
}

type mockHost struct {
	Commands []string
}

func (m mockHost) RunSSHCommand(cmd string) (string, error) {
	m.Commands = append(m.Commands, cmd)
	return "", nil
}

func TestStartCluster(t *testing.T) {
	h := mockHost{}
	err := StartCluster(h)
	if err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}
}

type mockHostError struct{}

func (m mockHostError) RunSSHCommand(cmd string) (string, error) {
	return "", fmt.Errorf("Error calling command: %s", cmd)
}

func TestStartClusterError(t *testing.T) {
	h := mockHostError{}
	err := StartCluster(h)
	if err == nil {
		t.Fatal("Error not thrown starting cluster.")
	}
}

func TestStartHostExists(t *testing.T) {
	api := &tests.MockAPI{}
	// Create an initial host.
	_, err := createHost(api)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}

	// Make sure the next call to Create will fail, to assert it doesn't get called again.
	api.CreateError = true
	if err := api.Create(&host.Host{}); err == nil {
		t.Fatal("api.Create did not fail, but should have.")
	}

	// This should pass without calling Create because the host exists already.
	h, err := StartHost(api)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != constants.MachineName {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
}

func TestStartStoppedHost(t *testing.T) {
	api := &tests.MockAPI{}
	// Create an initial host.
	h, err := createHost(api)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	d := tests.MockDriver{}
	h.Driver = &d
	d.CurrentState = state.Stopped

	h, err = StartHost(api)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != constants.MachineName {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}

	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
}

func TestStartHost(t *testing.T) {
	api := &tests.MockAPI{}

	h, err := StartHost(api)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != constants.MachineName {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}
	if exists, _ := api.Exists(h.Name); !exists {
		t.Fatal("Machine not saved.")
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
}

func TestStopHostError(t *testing.T) {
	api := &tests.MockAPI{}
	if err := StopHost(api); err == nil {
		t.Fatal("An error should be thrown when stopping non-existing machine.")
	}
}

func TestStopHost(t *testing.T) {
	api := &tests.MockAPI{}
	h, _ := createHost(api)
	if err := StopHost(api); err != nil {
		t.Fatal("An error should be thrown when stopping non-existing machine.")
	}
	if s, _ := h.Driver.GetState(); s != state.Stopped {
		t.Fatalf("Machine not stopped. Currently in state: %s", s)
	}

}

func TestMultiError(t *testing.T) {
	m := multiError{}

	m.Collect(fmt.Errorf("Error 1"))
	m.Collect(fmt.Errorf("Error 2"))

	err := m.ToError()
	expected := `Error 1
Error 2`
	if err.Error() != expected {
		t.Fatalf("%s != %s", err, expected)
	}

	m = multiError{}
	if err := m.ToError(); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestDeleteHost(t *testing.T) {
	api := &tests.MockAPI{}
	createHost(api)

	if err := DeleteHost(api); err != nil {
		t.Fatalf("Unexpected error deleting host: %s", err)
	}
}

func TestDeleteHostErrorDeletingVM(t *testing.T) {
	api := &tests.MockAPI{}
	h, _ := createHost(api)

	d := &tests.MockDriver{RemoveError: true}

	h.Driver = d

	if err := DeleteHost(api); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostErrorDeletingFiles(t *testing.T) {
	api := &tests.MockAPI{RemoveError: true}
	createHost(api)

	if err := DeleteHost(api); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostMultipleErrors(t *testing.T) {
	api := &tests.MockAPI{
		RemoveError: true,
	}
	h, _ := createHost(api)

	d := &tests.MockDriver{RemoveError: true}

	h.Driver = d

	err := DeleteHost(api)

	if err == nil {
		t.Fatal("Expected error deleting host, didn't get one.")
	}

	expectedErrors := []string{"Error removing minikubeVM", "Error deleting machine"}
	for _, expectedError := range expectedErrors {
		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error %s expected to contain: %s. ", err)
		}
	}
}

func TestGetHostStatus(t *testing.T) {
	api := &tests.MockAPI{}

	checkState := func(expected string) {
		s, err := GetHostStatus(api)
		if err != nil {
			t.Fatalf("Unexpected error getting status: %s", s)
		}
		if s != expected {
			t.Fatalf("Expected status: %s, got %s", s, expected)
		}
	}

	checkState("Does Not Exist")

	createHost(api)
	checkState(state.Running.String())

	StopHost(api)
	checkState(state.Stopped.String())
}
