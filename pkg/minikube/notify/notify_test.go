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

package notify

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
)

func TestCheckUpdateTimeAndVersionsLowerVersion(t *testing.T) {
	viper.Set(config.ReminderWaitPeriodInHours, 24)
	localVersion1, _ := semver.Make("1.0.0-dev")
	latestVersion4, _ := semver.Make("4.0.0-beta")
	//time.Time{} returns The zero value of type Time which is January 1, year 1, 00:00:00.000000000 UTC
	//24 hours should have elapsed from this time for the test
	if checkUpdateTimeAndVersions(time.Time{}, localVersion, latestVersion) == "" {
		t.Fatalf("Error starting cluster: %s", err)
	}
}

// Mock Host used for testing. When commands are run, the output from CommandOutput
// is used, if present. Then the output from Error is used, if present. Finally,
// "", nil is returned.
type mockHost struct {
	CommandOutput map[string]string
	Error         string
	Commands      map[string]int
}

func newMockHost() *mockHost {
	return &mockHost{
		CommandOutput: make(map[string]string),
		Commands:      make(map[string]int),
	}
}

func (m mockHost) RunSSHCommand(cmd string) (string, error) {
	m.Commands[cmd] = 1
	output, ok := m.CommandOutput[cmd]
	if ok {
		return output, nil
	}
	if m.Error != "" {
		return "", fmt.Errorf(m.Error)
	}
	return "", nil
}

func TestStartCluster(t *testing.T) {
	h := newMockHost()
	err := StartCluster(h)
	if err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}

	for _, cmd := range []string{stopCommand, GetStartCommand()} {
		if _, ok := h.Commands[cmd]; !ok {
			t.Fatalf("Expected command not run: %s. Commands run: %s", cmd, h.Commands)
		}
	}
}

func TestStartClusterError(t *testing.T) {
	h := newMockHost()
	h.Error = "error"

	err := StartCluster(h)
	if err == nil {
		t.Fatal("Error not thrown starting cluster.")
	}
}

func TestStartHostExists(t *testing.T) {
	api := tests.NewMockAPI()
	// Create an initial host.
	_, err := createHost(api, MachineConfig{})
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}

	// Make sure the next call to Create will fail, to assert it doesn't get called again.
	api.CreateError = true
	if err := api.Create(&host.Host{}); err == nil {
		t.Fatal("api.Create did not fail, but should have.")
	}

	// This should pass without calling Create because the host exists already.
	h, err := StartHost(api, MachineConfig{})
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
	api := tests.NewMockAPI()
	// Create an initial host.
	h, err := createHost(api, MachineConfig{})
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	d := tests.MockDriver{}
	h.Driver = &d
	d.CurrentState = state.Stopped

	h, err = StartHost(api, MachineConfig{})
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != constants.MachineName {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}

	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}

	if !api.SaveCalled {
		t.Fatalf("Machine must be saved after starting.")
	}
}

func TestStartHost(t *testing.T) {
	api := tests.NewMockAPI()

	h, err := StartHost(api, MachineConfig{})
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
	api := tests.NewMockAPI()
	if err := StopHost(api); err == nil {
		t.Fatal("An error should be thrown when stopping non-existing machine.")
	}
}

func TestStopHost(t *testing.T) {
	api := tests.NewMockAPI()
	h, _ := createHost(api, MachineConfig{})
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
	api := tests.NewMockAPI()
	createHost(api, MachineConfig{})

	if err := DeleteHost(api); err != nil {
		t.Fatalf("Unexpected error deleting host: %s", err)
	}
}

func TestDeleteHostErrorDeletingVM(t *testing.T) {
	api := tests.NewMockAPI()
	h, _ := createHost(api, MachineConfig{})

	d := &tests.MockDriver{RemoveError: true}

	h.Driver = d

	if err := DeleteHost(api); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostErrorDeletingFiles(t *testing.T) {
	api := tests.NewMockAPI()
	api.RemoveError = true
	createHost(api, MachineConfig{})

	if err := DeleteHost(api); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostMultipleErrors(t *testing.T) {
	api := tests.NewMockAPI()
	api.RemoveError = true
	h, _ := createHost(api, MachineConfig{})

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
	api := tests.NewMockAPI()

	checkState := func(expected string) {
		s, err := GetHostStatus(api)
		if err != nil {
			t.Fatalf("Unexpected error getting status: %s", err)
		}
		if s != expected {
			t.Fatalf("Expected status: %s, got %s", s, expected)
		}
	}

	checkState("Does Not Exist")

	createHost(api, MachineConfig{})
	checkState(state.Running.String())

	StopHost(api)
	checkState(state.Stopped.String())
}

func TestSetupCerts(t *testing.T) {
	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	if err := SetupCerts(d); err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}

	for _, cert := range certs {
		contents, _ := ioutil.ReadFile(cert)
		transferred := s.Transfers.Bytes()
		if !bytes.Contains(transferred, contents) {
			t.Fatalf("Certificate not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
		}
	}
}

func TestGetHostDockerEnv(t *testing.T) {
	api := tests.NewMockAPI()
	h, err := createHost(api, MachineConfig{})
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{
			IPAddress: "127.0.0.1",
		},
	}
	h.Driver = d

	envMap, err := GetHostDockerEnv(api)
	if err != nil {
		t.Fatalf("Unexpected error getting env: %s", err)
	}

	dockerEnvKeys := [...]string{
		"DOCKER_TLS_VERIFY",
		"DOCKER_HOST",
		"DOCKER_CERT_PATH",
	}
	for _, dockerEnvKey := range dockerEnvKeys {
		if _, hasKey := envMap[dockerEnvKey]; !hasKey {
			t.Fatalf("Expected envMap[\"%s\"] key to be defined", dockerEnvKey)
		}
	}
}

func TestHostGetLogs(t *testing.T) {
	api := tests.NewMockAPI()

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	api.Hosts[constants.MachineName] = &host.Host{Driver: d}

	if _, err := GetHostLogs(api); err != nil {
		t.Fatalf("Error getting host logs: %s", err)
	}

	if _, ok := s.Commands[logsCommand]; !ok {
		t.Fatalf("Expected command not run: %s", logsCommand)
	}
}

func TestCreateSSHShell(t *testing.T) {
	api := tests.NewMockAPI()

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port:         port,
		CurrentState: state.Running,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	api.Hosts[constants.MachineName] = &host.Host{Driver: d}

	cliArgs := []string{"exit"}
	if err := CreateSSHShell(api, cliArgs); err != nil {
		t.Fatalf("Error running ssh command: %s", err)
	}

	if s.HadASessionRequested != true {
		t.Fatalf("Expected ssh session to be run")
	}
}

type MockServiceGetter struct {
	services map[string]api.Service
}

func NewMockServiceGetter() *MockServiceGetter {
	return &MockServiceGetter{
		services: make(map[string]api.Service),
	}
}

func (mockServiceGetter *MockServiceGetter) Get(name string) (*api.Service, error) {
	service, ok := mockServiceGetter.services[name]
	if !ok {
		return nil, fmt.Errorf("Error getting %s service from mockServiceGetter", name)
	}
	return &service, nil
}

func TestGetDashboardURL(t *testing.T) {
	mockServiceGetter := NewMockServiceGetter()
	nodeport := api.ServicePort{
		NodePort: 1234,
	}
	mockDashboardService := api.Service{
		Spec: api.ServiceSpec{
			Ports: []api.ServicePort{nodeport},
		},
	}
	mockServiceGetter.services["kubernetes-dashboard"] = mockDashboardService

	port, err := getDashboardPortFromServiceGetter(mockServiceGetter)
	if err != nil {
		t.Fatalf("Error getting dashboard port from api: Error: ", err)
	}
	expected := 1234
	if port != expected {
		t.Fatalf("Error getting dashboard port from api: Expected: %s, Got: %s", port, expected)
	}

}
