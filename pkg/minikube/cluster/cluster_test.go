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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
)

var defaultMachineConfig = MachineConfig{VMDriver: constants.DefaultVMDriver,
	MinikubeISO: constants.DefaultIsoUrl}

func TestCreateHost(t *testing.T) {
	api := tests.NewMockAPI()

	exists, _ := api.Exists(constants.MachineName)
	if exists {
		t.Fatal("Machine already exists.")
	}
	_, err := createHost(api, defaultMachineConfig)
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

	found := false
	for _, driver := range constants.SupportedVMDrivers {
		if h.DriverName == driver {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Wrong driver name: %v. Should be virtualbox, vmwarefusion, kvm or xhyve.", h.DriverName)
	}
}

func TestStartCluster(t *testing.T) {
	h := tests.NewMockHost()
	ip, _ := h.Driver.GetIP()
	kubernetesConfig := KubernetesConfig{
		NodeIP: ip,
	}

	err := StartCluster(h, kubernetesConfig)
	if err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}

	for _, cmd := range []string{stopCommand, GetStartCommand(kubernetesConfig)} {
		if _, ok := h.Commands[cmd]; !ok {
			t.Fatalf("Expected command not run: %s. Commands run: %v", cmd, h.Commands)
		}
	}
}

func TestStartClusterError(t *testing.T) {
	h := tests.NewMockHost()
	h.Error = "error"
	ip, _ := h.Driver.GetIP()
	kubernetesConfig := KubernetesConfig{
		NodeIP: ip,
	}

	err := StartCluster(h, kubernetesConfig)
	if err == nil {
		t.Fatal("Error not thrown starting cluster.")
	}
}

func TestStartHostExists(t *testing.T) {
	api := tests.NewMockAPI()
	// Create an initial host.
	_, err := createHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}

	// Make sure the next call to Create will fail, to assert it doesn't get called again.
	api.CreateError = true
	if err := api.Create(&host.Host{}); err == nil {
		t.Fatal("api.Create did not fail, but should have.")
	}

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	// This should pass without calling Create because the host exists already.
	h, err := StartHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != constants.MachineName {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
	if !md.Provisioner.Provisioned {
		t.Fatalf("Expected provision to be called")
	}
}

func TestStartStoppedHost(t *testing.T) {
	api := tests.NewMockAPI()
	// Create an initial host.
	h, err := createHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	d := tests.MockDriver{}
	h.Driver = &d
	d.CurrentState = state.Stopped

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)
	h, err = StartHost(api, defaultMachineConfig)
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

	if !md.Provisioner.Provisioned {
		t.Fatalf("Expected provision to be called")
	}
}

func TestStartHost(t *testing.T) {
	api := tests.NewMockAPI()

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	h, err := StartHost(api, defaultMachineConfig)
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

	// Provision regenerates Docker certs. This happens automatically during create,
	// so we should only call it again if the host already exists.
	if md.Provisioner.Provisioned {
		t.Fatalf("Did not expect Provision to be called")
	}
}

func TestStartHostConfig(t *testing.T) {
	api := tests.NewMockAPI()

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	config := MachineConfig{
		VMDriver:  constants.DefaultVMDriver,
		DockerEnv: []string{"FOO=BAR"},
	}

	h, err := StartHost(api, config)
	if err != nil {
		t.Fatal("Error starting host.")
	}

	for i := range h.HostOptions.EngineOptions.Env {
		if h.HostOptions.EngineOptions.Env[i] != config.DockerEnv[i] {
			t.Fatal("Docker env variables were not set!")
		}
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
	h, _ := createHost(api, defaultMachineConfig)
	if err := StopHost(api); err != nil {
		t.Fatal("An error should be thrown when stopping non-existing machine.")
	}
	if s, _ := h.Driver.GetState(); s != state.Stopped {
		t.Fatalf("Machine not stopped. Currently in state: %s", s)
	}
}

func TestDeleteHost(t *testing.T) {
	api := tests.NewMockAPI()
	createHost(api, defaultMachineConfig)

	if err := DeleteHost(api); err != nil {
		t.Fatalf("Unexpected error deleting host: %s", err)
	}
}

func TestDeleteHostErrorDeletingVM(t *testing.T) {
	api := tests.NewMockAPI()
	h, _ := createHost(api, defaultMachineConfig)

	d := &tests.MockDriver{RemoveError: true}

	h.Driver = d

	if err := DeleteHost(api); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostErrorDeletingFiles(t *testing.T) {
	api := tests.NewMockAPI()
	api.RemoveError = true
	createHost(api, defaultMachineConfig)

	if err := DeleteHost(api); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostMultipleErrors(t *testing.T) {
	api := tests.NewMockAPI()
	api.RemoveError = true
	h, _ := createHost(api, defaultMachineConfig)

	d := &tests.MockDriver{RemoveError: true}

	h.Driver = d

	err := DeleteHost(api)

	if err == nil {
		t.Fatal("Expected error deleting host, didn't get one.")
	}

	expectedErrors := []string{"Error removing " + constants.MachineName, "Error deleting machine"}
	for _, expectedError := range expectedErrors {
		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error %s expected to contain: %s.", err, expectedError)
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

	createHost(api, defaultMachineConfig)
	checkState(state.Running.String())

	StopHost(api)
	checkState(state.Stopped.String())
}

func TestGetLocalkubeStatus(t *testing.T) {
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

	s.CommandToOutput = map[string]string{
		localkubeStatusCommand: state.Running.String(),
	}
	if _, err := GetLocalkubeStatus(api); err != nil {
		t.Fatalf("Error getting localkube status: %s", err)
	}

	s.CommandToOutput = map[string]string{
		localkubeStatusCommand: state.Stopped.String(),
	}
	if _, err := GetLocalkubeStatus(api); err != nil {
		t.Fatalf("Error getting localkube status: %s", err)
	}

	s.CommandToOutput = map[string]string{
		localkubeStatusCommand: "Bad Output",
	}
	if _, err := GetLocalkubeStatus(api); err == nil {
		t.Fatalf("Expected error in getting localkube status as ssh returned bad output")
	}
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
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	api := tests.NewMockAPI()
	h, err := createHost(api, defaultMachineConfig)
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
		return nil, errors.Errorf("Error getting %s service from mockServiceGetter", name)
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

	port, err := getServicePortFromServiceGetter(mockServiceGetter, "kubernetes-dashboard")
	if err != nil {
		t.Fatalf("Error getting dashboard port from api: Error: %s", err)
	}
	expected := 1234
	if port != expected {
		t.Fatalf("Error getting dashboard port from api: Expected: %d, Got: %d", port, expected)
	}

}

func TestGetServiceURLWithoutNodePort(t *testing.T) {
	mockServiceGetter := NewMockServiceGetter()
	mockDashboardService := api.Service{}
	mockServiceGetter.services["mock-service"] = mockDashboardService

	_, err := getServicePortFromServiceGetter(mockServiceGetter, "mock-service")
	if err == nil {
		t.Fatalf("Expected error getting service with no node port")
	}
}

func TestUpdateDefault(t *testing.T) {
	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	h := tests.NewMockHost()
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	kubernetesConfig := KubernetesConfig{
		KubernetesVersion: constants.DefaultKubernetesVersion,
	}

	if err := UpdateCluster(h, d, kubernetesConfig); err != nil {
		t.Fatalf("Error updating cluster: %s", err)
	}
	transferred := s.Transfers.Bytes()

	//test that kube-add memoryAssets are transferred properly
	for _, a := range memoryAssets {
		contents, _ := Asset(a.GetAssetName())
		if !bytes.Contains(transferred, contents) {
			t.Fatalf("File not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
		}
	}

	//test that localkube is transferred properly
	contents, _ := Asset("out/localkube")
	if !bytes.Contains(transferred, contents) {
		t.Fatalf("File not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
	}
}

var testLocalkubeBin = "hello"

type K8sVersionHandlerCorrect struct{}

func (h *K8sVersionHandlerCorrect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, testLocalkubeBin)
}

func TestUpdateKubernetesVersion(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	h := tests.NewMockHost()
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	handler := &K8sVersionHandlerCorrect{}
	server := httptest.NewServer(handler)

	kubernetesConfig := KubernetesConfig{
		KubernetesVersion: server.URL,
	}
	if err := UpdateCluster(h, d, kubernetesConfig); err != nil {
		t.Fatalf("Error updating cluster: %s", err)
	}
	transferred := s.Transfers.Bytes()

	//test that localkube is transferred properly
	contents := []byte(testLocalkubeBin)
	if !bytes.Contains(transferred, contents) {
		t.Fatalf("File not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
	}
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestIsLocalkubeCached(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	inputArr := [...]string{
		"v1.3.3",
		"1.3.0",
		"http://test-url.localkube.com/localkube-binary",
		"file:///test/dir/to/localkube-binary",
	}

	readCloser := nopCloser{}

	localkubeCacher := localkubeCacher{
		k8sConf: KubernetesConfig{},
	}

	inner := func(input string) {
		localkubeCacher.k8sConf = KubernetesConfig{
			KubernetesVersion: input,
		}
		if localkubeCacher.isLocalkubeCached() {
			t.Errorf("IsLocalKubeCached returned true even though %s was not cached",
				localkubeCacher.getLocalkubeCacheFilepath())
		}

		readCloser = nopCloser{bytes.NewBufferString("test-localkube-binary-data")}
		localkubeCacher.cacheLocalkube(readCloser)
		if !localkubeCacher.isLocalkubeCached() {
			t.Errorf("IsLocalKubeCached returned false even though %s was cached",
				localkubeCacher.getLocalkubeCacheFilepath())
		}

	}
	for _, input := range inputArr {
		inner(input)
	}
}

func TestIsIsoChecksumValid(t *testing.T) {
	tests := []struct {
		shouldMatch bool
		httpError   int
		expected    bool
	}{
		// SHA matches, no error.
		{true, 0, true},
		// SHA matches, HTTP error.
		{true, http.StatusNotFound, false},
		// SHA doesn't match.
		{false, 0, false},
		// SHA doesn't match, HTTP error.
		{false, http.StatusNotFound, false},
	}

	isoData := []byte("myIsoData")
	isoCheckSum := sha256.Sum256(isoData)
	for _, tc := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tc.httpError != 0 {
				w.WriteHeader(tc.httpError)
			}
			if tc.shouldMatch {
				io.WriteString(w, hex.EncodeToString(isoCheckSum[:]))
			} else {
				w.Write([]byte("badCheckSum"))
			}
		}))
		defer ts.Close()
		valid := isIsoChecksumValid(&isoData, ts.URL)
		if valid != tc.expected {
			t.Errorf("Expected isIsoChecksumValid to be %v, was %v", tc.expected, valid)
		}
	}
}

func TestUpdateCustomAddons(t *testing.T) {
	tempDir := tests.MakeTempDir()
	os.Mkdir(constants.MakeMiniPath("addons", "subdir"), 0777)
	defer os.RemoveAll(tempDir)

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	h := tests.NewMockHost()
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	//write a file into ~/.minikube/addons
	path := filepath.Join(constants.MakeMiniPath("addons"), "dir-addon.yaml")
	testContent1 := []byte("CUSTOM ADDON TEST STRING#1, In Addons Dir")
	err = ioutil.WriteFile(path, testContent1, 0644)
	if err != nil {
		t.Fatalf("Error writing custom addon: %s", err)
	}

	path = filepath.Join(constants.MakeMiniPath("addons", "subdir"), "subdir-addon.yaml")
	testContent2 := []byte("CUSTOM ADDON TEST STRING#2, In Addons SubDir")
	err = ioutil.WriteFile(path, testContent2, 0644)
	if err != nil {
		t.Fatalf("Error writing custom addon: %s", err)
	}

	//run update
	kubernetesConfig := KubernetesConfig{
		KubernetesVersion: constants.DefaultKubernetesVersion,
	}
	if err := UpdateCluster(h, d, kubernetesConfig); err != nil {
		t.Fatalf("Error updating cluster: %s", err)
	}
	transferred := s.Transfers.Bytes()

	//test that custom addons are transferred properly
	if !bytes.Contains(transferred, testContent1) {
		t.Fatalf("Custom addon not copied. Expected transfers to contain custom addon with content: %s. It was: %s", testContent1, transferred)
	}

	if !bytes.Contains(transferred, testContent2) {
		t.Fatalf("Custom addon not copied. Expected transfers to contain custom addon with content: %s. It was: %s", testContent2, transferred)
	}
}
